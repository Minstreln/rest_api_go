package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"reflect"
	"restapi/internal/models"
	"restapi/internal/repositories/sqlconnect"
	"restapi/pkg/utils"
	"strconv"
)

func GetStudentsHandler(w http.ResponseWriter, r *http.Request) {
	db, err := sqlconnect.ConnectDb()
	if err != nil {
		// http.Error(w, "Error connecting to database", http.StatusInternalServerError)
		utils.ErrorHandler(err, "Error retrieving data")
		return
	}
	defer db.Close()

	query := "SELECT id, first_name, last_name, email, class, subject FROM students WHERE 1=1"
	var args []interface{}

	query, args = utils.AddFilters(r, query, args)

	query = utils.AddSorting(r, query)

	rows, err := db.Query(query, args...)
	if err != nil {
		fmt.Println(err)
		// http.Error(w, "Database query error", http.StatusInternalServerError)
		utils.ErrorHandler(err, "Error retrieving data")
		return
	}
	defer rows.Close()

	studentList := make([]models.Student, 0)
	for rows.Next() {
		var student models.Student
		err := rows.Scan(&student.ID, &student.FirstName, &student.LastName, &student.Email, &student.Class)
		if err != nil {
			// http.Error(w, "Error scanning database results", http.StatusInternalServerError)
			utils.ErrorHandler(err, "Error retrieving data")
			return
		}
		studentList = append(studentList, student)
	}

	response := struct {
		Status string           `json:"status"`
		Count  int              `json:"count"`
		Data   []models.Student `json:"data"`
	}{
		Status: "success",
		Count:  len(studentList),
		Data:   studentList,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)

}

func GetOneStudentHandler(w http.ResponseWriter, r *http.Request) {
	db, err := sqlconnect.ConnectDb()
	if err != nil {
		// http.Error(w, "Error connecting to database", http.StatusInternalServerError)
		utils.ErrorHandler(err, "Error retrieving data")
		return
	}
	defer db.Close()

	idStr := r.PathValue("id")

	// Handle path parameter
	id, err := strconv.Atoi(idStr)
	if err != nil {
		fmt.Println(err)
		return
	}

	var student models.Student

	err = db.QueryRow("SELECT id, first_name, last_name, email, class, subject FROM students WHERE id = ?", id).Scan(&student.ID, &student.FirstName, &student.LastName, &student.Email, &student.Class)

	if err == sql.ErrNoRows {
		// http.Error(w, "Student not found", http.StatusNotFound)
		utils.ErrorHandler(err, "Student not found")
		return
	} else if err != nil {
		// http.Error(w, "Database query error", http.StatusInternalServerError)
		utils.ErrorHandler(err, "Error retrieving data")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(student)
}

// function for POST Student request handler
func AddStudentHandler(w http.ResponseWriter, r *http.Request) {
	db, err := sqlconnect.ConnectDb()
	if err != nil {
		// http.Error(w, "Error connecting to database", http.StatusInternalServerError)
		utils.ErrorHandler(err, "Error adding data")
		return
	}
	defer db.Close()

	var newStudents []models.Student
	var rawStudents []map[string]interface{}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error reading Request body", http.StatusInternalServerError)
		return
	}

	defer r.Body.Close()

	err = json.Unmarshal(body, &rawStudents)
	if err != nil {
		http.Error(w, "invalid Request body", http.StatusBadRequest)
		utils.ErrorHandler(err, "invalid Request body")
		return
	}

	fields := CheckFieldNames(models.Student{})

	allowedFields := make(map[string]struct{})
	for _, field := range fields {
		allowedFields[field] = struct{}{}
	}

	for _, student := range rawStudents {
		for key := range student {
			_, ok := allowedFields[key]
			if !ok {
				http.Error(w, "Unacceptable fields found in request. Only use allowed fields..", http.StatusBadRequest)
				return
			}

		}
	}

	err = json.Unmarshal(body, &newStudents)
	if err != nil {
		http.Error(w, "invalid Request body", http.StatusBadRequest)
		utils.ErrorHandler(err, "invalid Request body")
		return
	}

	for _, student := range newStudents {
		err := CheckBlankFields(student)
		if err != nil {
			// http.Error(w, "invalid Request body", http.StatusBadRequest)
			utils.ErrorHandler(err, "invalid Request body")
			return
		}
	}

	// stmt, err := db.Prepare("INSERT INTO students (first_name, last_name, email, class, subject) VALUES (?,?,?,?,?)")
	stmt, err := db.Prepare(utils.GenerateInsertQuery("students", models.Student{}))
	if err != nil {
		// http.Error(w, "Error preparing SQL query", http.StatusInternalServerError)
		utils.ErrorHandler(err, "Error adding data")
		return
	}
	defer stmt.Close()

	addedStudents := make([]models.Student, len(newStudents))
	for i, newStudent := range newStudents {
		// res, err := stmt.Exec(newStudent.FirstName, newStudent.LastName, newStudent.Email, newStudent.Class, newStudent.Subject)
		values := utils.GetStructValues(newStudent)
		res, err := stmt.Exec(values...)
		if err != nil {
			// http.Error(w, "Error inserting data into DB", http.StatusInternalServerError)
			utils.ErrorHandler(err, "Error adding data")
			return
		}
		lastID, err := res.LastInsertId()
		if err != nil {
			// http.Error(w, "Error getting last insert ID", http.StatusInternalServerError)
			utils.ErrorHandler(err, "Error adding data")
			return
		}
		newStudent.ID = int(lastID)
		addedStudents[i] = newStudent
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	response := struct {
		Status string           `json:"status"`
		Count  int              `json:"count"`
		Data   []models.Student `json:"data"`
	}{
		Status: "success",
		Count:  len(addedStudents),
		Data:   addedStudents,
	}

	json.NewEncoder(w).Encode(response)
}

// PUT /students/{id}
func UpdateStudentHandler(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid student ID", http.StatusBadRequest)
		utils.ErrorHandler(err, "Invalid student ID")
		return
	}

	var updatedStudent models.Student
	err = json.NewDecoder(r.Body).Decode(&updatedStudent)
	if err != nil {
		http.Error(w, "Invalid Request Payload", http.StatusBadRequest)
		return
	}

	db, err := sqlconnect.ConnectDb()
	if err != nil {
		// http.Error(w, "Unable to connect to database", http.StatusInternalServerError)
		utils.ErrorHandler(err, "Error updating data")
		return
	}
	defer db.Close()

	var existingStudent models.Student
	err = db.QueryRow("SELECT id, first_name, last_name, email, class, subject FROM students WHERE id = ?", id).Scan(&existingStudent.ID, &existingStudent.FirstName, &existingStudent.LastName, &existingStudent.Email, &existingStudent.Class)
	if err == sql.ErrNoRows {
		http.Error(w, "Student not found", http.StatusNotFound)
		utils.ErrorHandler(err, "Student not found")
		return
	} else if err != nil {
		// http.Error(w, "Unable to retrieve data", http.StatusInternalServerError)
		utils.ErrorHandler(err, "Error updating data")
		return
	}

	updatedStudent.ID = existingStudent.ID
	_, err = db.Exec("UPDATE students SET first_name = ?, last_name = ?, email = ?, class = ?, subject = ? WHERE id = ?", updatedStudent.FirstName, updatedStudent.LastName, updatedStudent.Email, updatedStudent.Class, updatedStudent.ID)
	if err != nil {
		// http.Error(w, "Unable to update student", http.StatusInternalServerError)
		utils.ErrorHandler(err, "Error updating data")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updatedStudent)

}

// PATCH FOR MULTIPLE ENTRIES /students
func PatchStudentsHandler(w http.ResponseWriter, r *http.Request) {
	db, err := sqlconnect.ConnectDb()
	if err != nil {
		// http.Error(w, "Unable to connect to database", http.StatusInternalServerError)
		utils.ErrorHandler(err, "Error updating data")
		return
	}
	defer db.Close()

	var updates []map[string]interface{}
	err = json.NewDecoder(r.Body).Decode(&updates)
	if err != nil {
		// http.Error(w, "Invalid request payload", http.StatusBadRequest)
		utils.ErrorHandler(err, "Invalid request payload")
		return
	}

	// start transaction
	tx, err := db.Begin()
	if err != nil {
		// http.Error(w, "Error starting transaction", http.StatusInternalServerError)
		utils.ErrorHandler(err, "Error updating data")
		return
	}

	for _, update := range updates {
		idStr, ok := update["id"].(string)
		if !ok {
			tx.Rollback()
			// http.Error(w, "Invalid student ID in update", http.StatusBadRequest)
			utils.ErrorHandler(err, "Invalid student ID")
			return
		}

		id, err := strconv.Atoi(idStr)
		if err != nil {
			tx.Rollback()
			// http.Error(w, "Error converting ID to int", http.StatusInternalServerError)
			utils.ErrorHandler(err, "Error updating data")
			return
		}

		var studentFromDb models.Student
		err = db.QueryRow("SELECT id, first_name, last_name, email, class, subject FROM students WHERE id = ?", id).Scan(&studentFromDb.ID, &studentFromDb.FirstName, &studentFromDb.LastName, &studentFromDb.Email, &studentFromDb.Class)
		if err != nil {
			tx.Rollback()
			if err == sql.ErrNoRows {
				http.Error(w, "Student not found", http.StatusNotFound)
				utils.ErrorHandler(err, "Student not found")
			}
			// http.Error(w, "Error retrieving Student", http.StatusInternalServerError)
			utils.ErrorHandler(err, "Error updating data")
			return
		}

		// apply updates using reflect
		studentVal := reflect.ValueOf(&studentFromDb).Elem()
		studentType := studentVal.Type()

		for k, v := range update {
			if k == "id" {
				continue // skip updating the id field
			}
			for i := 0; i < studentVal.NumField(); i++ {
				field := studentType.Field(i)
				if field.Tag.Get("json") == k+",omitempty" {
					fieldVal := studentVal.Field(i)
					if fieldVal.CanSet() {
						val := reflect.ValueOf(v)
						if val.Type().ConvertibleTo(fieldVal.Type()) {
							fieldVal.Set(val.Convert(fieldVal.Type()))
						} else {
							tx.Rollback()
							log.Printf("Cannot convert %v to %v", val.Type(), fieldVal.Type())
							return
						}
					}
					break
				}
			}
		}

		_, err = tx.Exec("UPDATE students SET first_name = ?, last_name = ?, email = ?, class = ?, subject = ? WHERE id = ?", studentFromDb.FirstName, studentFromDb.LastName, studentFromDb.Email, studentFromDb.Class, studentFromDb.ID)
		if err != nil {
			tx.Rollback()
			// http.Error(w, "Error updating student", http.StatusInternalServerError)
			utils.ErrorHandler(err, "Error updating data")
			return
		}
	}

	// commit the transaction
	err = tx.Commit()
	if err != nil {
		// http.Error(w, "Error committing transaction", http.StatusInternalServerError)
		utils.ErrorHandler(err, "Error updating data")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// PATCH /students/{id}
func PatchOneStudentHandler(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		// http.Error(w, "Invalid Student ID", http.StatusBadRequest)
		utils.ErrorHandler(err, "Invalid Student ID")
		return
	}

	var updates map[string]interface{}
	err = json.NewDecoder(r.Body).Decode(&updates)
	if err != nil {
		// http.Error(w, "Invalid Request Payload", http.StatusBadRequest)
		utils.ErrorHandler(err, "Invalid Request payload")
		return
	}

	db, err := sqlconnect.ConnectDb()
	if err != nil {
		// http.Error(w, "Unable to connect to database", http.StatusInternalServerError)
		utils.ErrorHandler(err, "Error updating data")
		return
	}
	defer db.Close()

	var existingStudent models.Student
	err = db.QueryRow("SELECT id, first_name, last_name, email, class, subject FROM students WHERE id = ?", id).Scan(&existingStudent.ID, &existingStudent.FirstName, &existingStudent.LastName, &existingStudent.Email, &existingStudent.Class)
	if err == sql.ErrNoRows {
		// http.Error(w, "Student not found", http.StatusNotFound)
		utils.ErrorHandler(err, "Student not found")
		return
	} else if err != nil {
		// http.Error(w, "Unable to retrieve data", http.StatusInternalServerError)
		utils.ErrorHandler(err, "Error updating data")
		return
	}

	// apply updates using reflect
	studentVal := reflect.ValueOf(&existingStudent).Elem()
	studentType := studentVal.Type()

	for k, v := range updates {
		for i := 0; i < studentVal.NumField(); i++ {
			field := studentType.Field(i)
			field.Tag.Get("json")
			if field.Tag.Get("json") == k+",omitempty" {
				if studentVal.Field(i).CanSet() {
					studentVal.Field(i).Set(reflect.ValueOf(v).Convert(studentVal.Field(i).Type()))
				}
			}
		}
	}

	_, err = db.Exec("UPDATE students SET first_name = ?, last_name = ?, email = ?, class = ?, subject = ? WHERE id = ?", existingStudent.FirstName, existingStudent.LastName, existingStudent.Email, existingStudent.Class, existingStudent.ID)
	if err != nil {
		// http.Error(w, "Unable to update student", http.StatusInternalServerError)
		utils.ErrorHandler(err, "Error updating data")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(existingStudent)

}

// DELETE /students/{id}
func DeleteOneStudentHandler(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		// http.Error(w, "Invalid Student ID", http.StatusBadRequest)
		utils.ErrorHandler(err, "Invalid Student ID")
		return
	}

	db, err := sqlconnect.ConnectDb()
	if err != nil {
		// http.Error(w, "Unable to connect to database", http.StatusInternalServerError)
		utils.ErrorHandler(err, "Error deleting data")
		return
	}
	defer db.Close()

	res, err := db.Exec("DELETE FROM students WHERE id = ?", id)
	if err != nil {
		// http.Error(w, "Error deleting student", http.StatusInternalServerError)
		utils.ErrorHandler(err, "Error deleting data")
		return
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		// http.Error(w, "Error retrieving delete result", http.StatusInternalServerError)
		utils.ErrorHandler(err, "Error deleting data")
		return
	}

	if rowsAffected == 0 {
		// http.Error(w, "Student not found", http.StatusNotFound)
		utils.ErrorHandler(err, "Student not found")
		return
	}

	w.WriteHeader(http.StatusNoContent)

}

// DELETE MULTIPLE STUDENTS /studentrs
func DeleteStudentsHandler(w http.ResponseWriter, r *http.Request) {
	db, err := sqlconnect.ConnectDb()
	if err != nil {
		// http.Error(w, "Unable to connect to database", http.StatusInternalServerError)
		utils.ErrorHandler(err, "Error deleting data")
		return
	}
	defer db.Close()

	var ids []int
	err = json.NewDecoder(r.Body).Decode(&ids)
	if err != nil {
		// http.Error(w, "Invalid request payload", http.StatusBadRequest)
		utils.ErrorHandler(err, "Invalid request payload")
		return
	}

	tx, err := db.Begin()
	if err != nil {
		// http.Error(w, "Error starting transaction", http.StatusInternalServerError)
		utils.ErrorHandler(err, "Error deleting data")
		return
	}

	stmt, err := tx.Prepare("DELETE FROM students WHERE id = ?")
	if err != nil {
		tx.Rollback()
		// http.Error(w, "Error preparing delete statement", http.StatusInternalServerError)
		utils.ErrorHandler(err, "Error deleting data")
		return
	}
	defer stmt.Close()

	deletedIds := []int{}
	for _, id := range ids {
		res, err := stmt.Exec(id)
		if err != nil {
			tx.Rollback()

			// http.Error(w, "Error Deleting Student", http.StatusInternalServerError)
			utils.ErrorHandler(err, "Error deleting data")
			return
		}

		rowsAffected, err := res.RowsAffected()
		if err != nil {
			tx.Rollback()
			// http.Error(w, "Error retrieving deleted result", http.StatusInternalServerError)
			utils.ErrorHandler(err, "Error deleting data")
			return
		}

		// if student was deleted then add the deleted id to deletedIDs slice
		if rowsAffected > 0 {
			deletedIds = append(deletedIds, id)
		}

		if rowsAffected < 1 {
			tx.Rollback()
			// http.Error(w, fmt.Sprintf("ID %v does not exist", id), http.StatusInternalServerError)
			utils.ErrorHandler(err, fmt.Sprintf("ID %v does not exist", id))
			return
		}
	}

	// commit
	err = tx.Commit()
	if err != nil {
		// http.Error(w, "Error committing transaction", http.StatusInternalServerError)
		utils.ErrorHandler(err, "Error deleting data")
		return
	}

	if len(deletedIds) < 1 {
		// http.Error(w, "IDs do not exists", http.StatusBadRequest)
		utils.ErrorHandler(err, "IDs do not exist")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	response := struct {
		Status     string `json:"status"`
		DeletedIDs []int  `json:"deleted_ids"`
	}{
		Status:     "Students successfully deleted",
		DeletedIDs: deletedIds,
	}

	json.NewEncoder(w).Encode(response)
}
