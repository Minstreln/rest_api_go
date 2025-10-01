package sqlconnect

import (
	"database/sql"
	"fmt"
	"net/http"
	"restapi/internal/models"
	"strings"
)

func isValidSortOrder(order string) bool {
	return order == "asc" || order == "desc"
}

func isValidSortField(field string) bool {
	validFields := map[string]bool{
		"first_name": true,
		"last_name":  true,
		"email":      true,
		"class":      true,
		"subject":    true,
	}
	return validFields[field]
}

func addSorting(r *http.Request, query string) string {
	sortParams := r.URL.Query()["sortby"]
	if len(sortParams) > 0 {
		query += " ORDER BY"
		for i, param := range sortParams {
			parts := strings.Split(param, ":")
			if len(parts) != 2 {
				continue
			}
			field, order := parts[0], parts[1]
			if !isValidSortField(field) || !isValidSortOrder(order) {
				continue
			}
			if i > 0 {
				query += ","
			}
			query += " " + field + " " + order
		}
	}
	return query
}

func addFilters(r *http.Request, query string, args []interface{}) (string, []interface{}) {
	params := map[string]string{
		"first_name": "first_name",
		"last_name":  "last_name",
		"email":      "email",
		"class":      "class",
		"subject":    "subject",
	}

	for param, dbField := range params {
		value := r.URL.Query().Get(param)
		if value != "" {
			query += " AND " + dbField + " = ?"
			args = append(args, value)
		}
	}
	return query, args
}

func GetTeachersDbHandler(r *http.Request) ([]models.Teacher, bool) {
	db, err := ConnectDb()
	if err != nil {
		// http.Error(w, "Error connecting to database", http.StatusInternalServerError)
		return nil, true
	}
	defer db.Close()

	query := "SELECT id, first_name, last_name, email, class, subject FROM teachers WHERE 1=1"
	var args []interface{}

	query, args = addFilters(r, query, args)

	query = addSorting(r, query)

	rows, err := db.Query(query, args...)
	if err != nil {
		fmt.Println(err)
		// http.Error(w, "Database query error", http.StatusInternalServerError)
		return nil, true
	}
	defer rows.Close()

	teacherList := make([]models.Teacher, 0)
	for rows.Next() {
		var teacher models.Teacher
		err := rows.Scan(&teacher.ID, &teacher.FirstName, &teacher.LastName, &teacher.Email, &teacher.Class, &teacher.Subject)
		if err != nil {
			// http.Error(w, "Error scanning database results", http.StatusInternalServerError)
			return nil, true
		}
		teacherList = append(teacherList, teacher)
	}
	return teacherList, false
}

func GetOneTeacherDbHandler(w http.ResponseWriter, id int) (models.Teacher, bool) {
	db, err := ConnectDb()
	if err != nil {
		http.Error(w, "Error connecting to database", http.StatusInternalServerError)
		return models.Teacher{}, true
	}
	defer db.Close()

	var teacher models.Teacher

	err = db.QueryRow("SELECT id, first_name, last_name, email, class, subject FROM teachers WHERE id = ?", id).Scan(&teacher.ID, &teacher.FirstName, &teacher.LastName, &teacher.Email, &teacher.Class, &teacher.Subject)

	if err == sql.ErrNoRows {
		http.Error(w, "Teacher not found", http.StatusNotFound)
		return models.Teacher{}, true
	} else if err != nil {
		http.Error(w, "Database query error", http.StatusInternalServerError)
		return models.Teacher{}, true
	}
	return teacher, false
}
