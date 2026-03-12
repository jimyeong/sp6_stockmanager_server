package v1

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/jimyeongjung/owlverload_api/models"
)

type Tag struct {
	ID      string `json:"id"`
	TagName string `json:"tag_name"`
}

// GetTagByID retrieves a tag by its ID
func GetTagByID(id string) (Tag, error) {
	if id == "" {
		return Tag{}, fmt.Errorf("empty tag ID")
	}
	db := models.GetDBInstance(models.GetDBConfig())
	var t Tag
	query := "SELECT id, name as tag_name FROM tags WHERE id = ?"
	err := db.QueryRow(query, id).Scan(&t.ID, &t.TagName)
	if err != nil {
		if err == sql.ErrNoRows {
			return Tag{}, errors.New("tag not found")
		}
		return Tag{}, err
	}
	return t, nil
}

// GetTagsForProduct retrieves all tags associated with a product (item)
func GetTagsForProduct(productID string) ([]Tag, error) {
	if productID == "" {
		return nil, fmt.Errorf("empty product ID")
	}
	db := models.GetDBInstance(models.GetDBConfig())
	query := `SELECT t.id, t.name as tag_name
		FROM tags t
		JOIN item_tags it ON t.id = it.tag_id
		WHERE it.item_id = ?`
	rows, err := db.Query(query, productID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var tags []Tag
	for rows.Next() {
		var t Tag
		err := rows.Scan(&t.ID, &t.TagName)
		if err != nil {
			return nil, err
		}
		tags = append(tags, t)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return tags, nil
}

// GetAllTags retrieves all tags from the database
func GetAllTags() ([]Tag, error) {
	db := models.GetDBInstance(models.GetDBConfig())
	query := "SELECT id, name as tag_name FROM tags"
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var tags []Tag
	for rows.Next() {
		var t Tag
		err := rows.Scan(&t.ID, &t.TagName)
		if err != nil {
			return nil, err
		}
		tags = append(tags, t)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return tags, nil
}
