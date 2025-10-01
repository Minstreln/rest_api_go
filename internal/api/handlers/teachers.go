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

func GetTeachersHandler(w http.ResponseWriter, r *http.Request) {
	db, err := sqlconnect.ConnectDb()
	if err != nil {
		// http.Error(w, "Error connecting to database", http.StatusInternalServerError)
		utils.ErrorHandler(err, "Error retrieving data")
		return
	}
	defer db.Close()

	query := "SELECT id, first_name, last_name, email, class, subject FROM teachers WHERE 1=1"
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

	teacherList := make([]models.Teacher, 0)
	for rows.Next() {
		var teacher models.Teacher
		err := rows.Scan(&teacher.ID, &teacher.FirstName, &teacher.LastName, &teacher.Email, &teacher.Class, &teacher.Subject)
		if err != nil {
			// http.Error(w, "Error scanning database results", http.StatusInternalServerError)
			utils.ErrorHandler(err, "Error retrieving data")
			return
		}
		teacherList = append(teacherList, teacher)
	}

	response := struct {
		Status string           `json:"status"`
		Count  int              `json:"count"`
		Data   []models.Teacher `json:"data"`
	}{
		Status: "success",
		Count:  len(teacherList),
		Data:   teacherList,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)

}

func GetOneTeacherHandler(w http.ResponseWriter, r *http.Request) {
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

	var teacher models.Teacher

	err = db.QueryRow("SELECT id, first_name, last_name, email, class, subject FROM teachers WHERE id = ?", id).Scan(&teacher.ID, &teacher.FirstName, &teacher.LastName, &teacher.Email, &teacher.Class, &teacher.Subject)

	if err == sql.ErrNoRows {
		// http.Error(w, "Teacher not found", http.StatusNotFound)
		utils.ErrorHandler(err, "Teacher not found")
		return
	} else if err != nil {
		// http.Error(w, "Database query error", http.StatusInternalServerError)
		utils.ErrorHandler(err, "Error retrieving data")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(teacher)
}

// function for POST Teacher request handler
func AddTeacherHandler(w http.ResponseWriter, r *http.Request) {
	db, err := sqlconnect.ConnectDb()
	if err != nil {
		// http.Error(w, "Error connecting to database", http.StatusInternalServerError)
		utils.ErrorHandler(err, "Error adding data")
		return
	}
	defer db.Close()

	var newTeachers []models.Teacher
	var rawTeachers []map[string]interface{}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error reading Request body", http.StatusInternalServerError)
		return
	}

	defer r.Body.Close()

	err = json.Unmarshal(body, &rawTeachers)
	if err != nil {
		http.Error(w, "invalid Request body", http.StatusBadRequest)
		utils.ErrorHandler(err, "invalid Request body")
		return
	}

	fields := CheckFieldNames(models.Teacher{})

	allowedFields := make(map[string]struct{})
	for _, field := range fields {
		allowedFields[field] = struct{}{}
	}

	for _, teacher := range rawTeachers {
		for key := range teacher {
			_, ok := allowedFields[key]
			if !ok {
				http.Error(w, "Unacceptable fields found in request. Only use allowed fields..", http.StatusBadRequest)
				return
			}

		}
	}

	err = json.Unmarshal(body, &newTeachers)
	if err != nil {
		http.Error(w, "invalid Request body", http.StatusBadRequest)
		utils.ErrorHandler(err, "invalid Request body")
		return
	}

	for _, teacher := range newTeachers {
		err := CheckBlankFields(teacher)
		if err != nil {
			// http.Error(w, "invalid Request body", http.StatusBadRequest)
			utils.ErrorHandler(err, "invalid Request body")
			return
		}
	}

	// stmt, err := db.Prepare("INSERT INTO teachers (first_name, last_name, email, class, subject) VALUES (?,?,?,?,?)")
	stmt, err := db.Prepare(utils.GenerateInsertQuery("teachers", models.Teacher{}))
	if err != nil {
		// http.Error(w, "Error preparing SQL query", http.StatusInternalServerError)
		utils.ErrorHandler(err, "Error adding data")
		return
	}
	defer stmt.Close()

	addedTeachers := make([]models.Teacher, len(newTeachers))
	for i, newTeacher := range newTeachers {
		// res, err := stmt.Exec(newTeacher.FirstName, newTeacher.LastName, newTeacher.Email, newTeacher.Class, newTeacher.Subject)
		values := utils.GetStructValues(newTeacher)
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
		newTeacher.ID = int(lastID)
		addedTeachers[i] = newTeacher
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	response := struct {
		Status string           `json:"status"`
		Count  int              `json:"count"`
		Data   []models.Teacher `json:"data"`
	}{
		Status: "success",
		Count:  len(addedTeachers),
		Data:   addedTeachers,
	}

	json.NewEncoder(w).Encode(response)
}

// PUT /teachers/{id}
func UpdateTeacherHandler(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid Teacher ID", http.StatusBadRequest)
		utils.ErrorHandler(err, "Invalid teacher ID")
		return
	}

	var updatedTeader models.Teacher
	err = json.NewDecoder(r.Body).Decode(&updatedTeader)
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

	var existingTeacher models.Teacher
	err = db.QueryRow("SELECT id, first_name, last_name, email, class, subject FROM teachers WHERE id = ?", id).Scan(&existingTeacher.ID, &existingTeacher.FirstName, &existingTeacher.LastName, &existingTeacher.Email, &existingTeacher.Class, &existingTeacher.Subject)
	if err == sql.ErrNoRows {
		http.Error(w, "Teacher not found", http.StatusNotFound)
		utils.ErrorHandler(err, "Teacher not found")
		return
	} else if err != nil {
		// http.Error(w, "Unable to retrieve data", http.StatusInternalServerError)
		utils.ErrorHandler(err, "Error updating data")
		return
	}

	updatedTeader.ID = existingTeacher.ID
	_, err = db.Exec("UPDATE teachers SET first_name = ?, last_name = ?, email = ?, class = ?, subject = ? WHERE id = ?", updatedTeader.FirstName, updatedTeader.LastName, updatedTeader.Email, updatedTeader.Class, updatedTeader.Subject, updatedTeader.ID)
	if err != nil {
		// http.Error(w, "Unable to update teacher", http.StatusInternalServerError)
		utils.ErrorHandler(err, "Error updating data")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updatedTeader)

}

// PATCH FOR MULTIPLE ENTRIES /teachers
func PatchTeachersHandler(w http.ResponseWriter, r *http.Request) {
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
			// http.Error(w, "Invalid teacher ID in update", http.StatusBadRequest)
			utils.ErrorHandler(err, "Invalid teacher ID")
			return
		}

		id, err := strconv.Atoi(idStr)
		if err != nil {
			tx.Rollback()
			// http.Error(w, "Error converting ID to int", http.StatusInternalServerError)
			utils.ErrorHandler(err, "Error updating data")
			return
		}

		var teacherFromDb models.Teacher
		err = db.QueryRow("SELECT id, first_name, last_name, email, class, subject FROM teachers WHERE id = ?", id).Scan(&teacherFromDb.ID, &teacherFromDb.FirstName, &teacherFromDb.LastName, &teacherFromDb.Email, &teacherFromDb.Class, &teacherFromDb.Subject)
		if err != nil {
			tx.Rollback()
			if err == sql.ErrNoRows {
				http.Error(w, "Teacher not found", http.StatusNotFound)
				utils.ErrorHandler(err, "Teacher not found")
			}
			// http.Error(w, "Error retrieving Teacher", http.StatusInternalServerError)
			utils.ErrorHandler(err, "Error updating data")
			return
		}

		// apply updates using reflect
		teacherVal := reflect.ValueOf(&teacherFromDb).Elem()
		teacherType := teacherVal.Type()

		for k, v := range update {
			if k == "id" {
				continue // skip updating the id field
			}
			for i := 0; i < teacherVal.NumField(); i++ {
				field := teacherType.Field(i)
				if field.Tag.Get("json") == k+",omitempty" {
					fieldVal := teacherVal.Field(i)
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

		_, err = tx.Exec("UPDATE teachers SET first_name = ?, last_name = ?, email = ?, class = ?, subject = ? WHERE id = ?", teacherFromDb.FirstName, teacherFromDb.LastName, teacherFromDb.Email, teacherFromDb.Class, teacherFromDb.Subject, teacherFromDb.ID)
		if err != nil {
			tx.Rollback()
			// http.Error(w, "Error updating teacher", http.StatusInternalServerError)
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

// PATCH /teachers/{id}
func PatchOneTeacherHandler(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		// http.Error(w, "Invalid Teacher ID", http.StatusBadRequest)
		utils.ErrorHandler(err, "Invalid Teacher ID")
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

	var existingTeacher models.Teacher
	err = db.QueryRow("SELECT id, first_name, last_name, email, class, subject FROM teachers WHERE id = ?", id).Scan(&existingTeacher.ID, &existingTeacher.FirstName, &existingTeacher.LastName, &existingTeacher.Email, &existingTeacher.Class, &existingTeacher.Subject)
	if err == sql.ErrNoRows {
		// http.Error(w, "Teacher not found", http.StatusNotFound)
		utils.ErrorHandler(err, "Teacher not found")
		return
	} else if err != nil {
		// http.Error(w, "Unable to retrieve data", http.StatusInternalServerError)
		utils.ErrorHandler(err, "Error updating data")
		return
	}

	// apply updates using reflect
	teacherVal := reflect.ValueOf(&existingTeacher).Elem()
	teacherType := teacherVal.Type()

	for k, v := range updates {
		for i := 0; i < teacherVal.NumField(); i++ {
			field := teacherType.Field(i)
			field.Tag.Get("json")
			if field.Tag.Get("json") == k+",omitempty" {
				if teacherVal.Field(i).CanSet() {
					teacherVal.Field(i).Set(reflect.ValueOf(v).Convert(teacherVal.Field(i).Type()))
				}
			}
		}
	}

	_, err = db.Exec("UPDATE teachers SET first_name = ?, last_name = ?, email = ?, class = ?, subject = ? WHERE id = ?", existingTeacher.FirstName, existingTeacher.LastName, existingTeacher.Email, existingTeacher.Class, existingTeacher.Subject, existingTeacher.ID)
	if err != nil {
		// http.Error(w, "Unable to update teacher", http.StatusInternalServerError)
		utils.ErrorHandler(err, "Error updating data")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(existingTeacher)

}

// DELETE /teachers/{id}
func DeleteOneTeacherHandler(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		// http.Error(w, "Invalid Teacher ID", http.StatusBadRequest)
		utils.ErrorHandler(err, "Invalid Teacher ID")
		return
	}

	db, err := sqlconnect.ConnectDb()
	if err != nil {
		// http.Error(w, "Unable to connect to database", http.StatusInternalServerError)
		utils.ErrorHandler(err, "Error deleting data")
		return
	}
	defer db.Close()

	res, err := db.Exec("DELETE FROM teachers WHERE id = ?", id)
	if err != nil {
		// http.Error(w, "Error deleting teacher", http.StatusInternalServerError)
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
		// http.Error(w, "Teacher not found", http.StatusNotFound)
		utils.ErrorHandler(err, "Teacher not found")
		return
	}

	w.WriteHeader(http.StatusNoContent)

}

// DELETE MULTIPLE TEACHERS /teachers
func DeleteTeachersHandler(w http.ResponseWriter, r *http.Request) {
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

	stmt, err := tx.Prepare("DELETE FROM teachers WHERE id = ?")
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

			// http.Error(w, "Error Deleting Teacher", http.StatusInternalServerError)
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

		// if teacher was deleted then add the deleted id to deletedIDs slice
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
		Status:     "Teachers successfully deleted",
		DeletedIDs: deletedIds,
	}

	json.NewEncoder(w).Encode(response)
}
