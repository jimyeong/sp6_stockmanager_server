package models

import (
	"fmt"
	"time"
)

// Tag represents a product tag used for categorization and recommendations
type Tag struct {
	ID      string `json:"id"`
	TagName string `json:"tag_name"`
	// tagNa  string    `json:"category"` // For grouping related tags
	// CreatedAt time.Time `json:"createdAt"`
}

// ItemTag represents the many-to-many relationship between items and tags
type ItemTag struct {
	ItemID    string    `json:"itemId"`
	TagID     string    `json:"tagId"`
	CreatedAt time.Time `json:"createdAt"`
}

// GetAllTags retrieves all tags from the database
func GetAllTags() ([]Tag, error) {
	fmt.Println("---GETALLTAGS---")
	db := GetDBInstance(GetDBConfig())
	var tags []Tag

	query := "SELECT id, name as tag_name FROM tags"
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var tag Tag
		err := rows.Scan(
			&tag.ID,
			&tag.TagName,
		)
		if err != nil {
			return nil, err
		}
		tags = append(tags, tag)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return tags, nil
}

// GetTagsByCategory retrieves all tags in a specific category
func GetTagsByCategory(category string) ([]Tag, error) {
	fmt.Println("---GETTAGSBYCATEGORY---", category)
	db := GetDBInstance(GetDBConfig())
	var tags []Tag

	query := "SELECT id, name, category, created_at FROM tags WHERE category = ?"
	rows, err := db.Query(query, category)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var tag Tag
		err := rows.Scan(
			&tag.ID,
			&tag.TagName,
			// &tag.Category,
			// &tag.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		tags = append(tags, tag)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return tags, nil
}

// GetItemsByTags retrieves items that have any of the given tags
func GetItemsByTags(tagIDs []string) ([]Item, error) {
	fmt.Println("---GETITEMSBYTAGS---", tagIDs)
	if len(tagIDs) == 0 {
		return nil, fmt.Errorf("no tags provided")
	}

	db := GetDBInstance(GetDBConfig())
	var items []Item

	// Create placeholders for SQL query
	placeholders := ""
	args := make([]interface{}, len(tagIDs))

	for i, tagID := range tagIDs {
		if i > 0 {
			placeholders += ", "
		}
		placeholders += "?"
		args[i] = tagID
	}

	// Query to get items that have any of the given tags
	query := `
	SELECT DISTINCT i.item_id, IFNULL(i.code, ''), IFNULL(i.barcode, ''), IFNULL(i.box_barcode, ''), IFNULL(i.price, 0), IFNULL(i.box_price, 0), IFNULL(i.name, ''), 
	IFNULL(i.type, ''), IFNULL(i.available_for_order, 0), IFNULL(i.image_path, ''), i.created_at
	FROM items i
	JOIN item_tags it ON i.item_id = it.item_id
	WHERE it.tag_id IN (` + placeholders + `)
	ORDER BY i.created_at DESC
	LIMIT 50`

	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var item Item
		err := rows.Scan(
			&item.ID,
			&item.Code,
			&item.BarCode,
			&item.BoxBarcode,
			&item.Name,
			&item.Type,
			&item.AvailableForOrder,
			&item.ImagePath,
			&item.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return items, nil
}

// GetTagsForItem retrieves all tags associated with a specific item
func GetTagsForItem(itemID string) ([]Tag, error) {
	fmt.Println("---GETTAGSFORITEM---", itemID)
	db := GetDBInstance(GetDBConfig())
	var tags []Tag

	query := `
	SELECT t.id, t.name
	FROM tags t
	JOIN item_tags it ON t.id = it.tag_id
	WHERE it.item_id = ?`

	rows, err := db.Query(query, itemID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var tag Tag
		err := rows.Scan(
			&tag.ID,
			&tag.TagName,
			// &tag.Category,
			// &tag.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		tags = append(tags, tag)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return tags, nil
}

// AssociateItemWithTags links an item with multiple tags
func AssociateItemWithTags(itemID string, tagIDs []string) error {
	fmt.Println("---ASSOCIATEITEMWITHTAGS---", itemID, tagIDs)
	if len(tagIDs) == 0 {
		return nil // Nothing to do
	}

	db := GetDBInstance(GetDBConfig())
	tx, err := db.Begin()
	if err != nil {
		return err
	}

	// Prepare the statement for insertion
	stmt, err := tx.Prepare("INSERT IGNORE INTO item_tags (item_id, tag_id, created_at) VALUES (?, ?, ?)")
	if err != nil {
		tx.Rollback()
		return err
	}
	defer stmt.Close()

	now := time.Now()
	for _, tagID := range tagIDs {
		_, err := stmt.Exec(itemID, tagID, now)
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit()
}

func UpdateTagsForItem(itemID string, tagIDs []string) error {
	fmt.Println("---UPDATEITEMTAGS---", itemID, tagIDs)
	if itemID == "" {
		return fmt.Errorf("itemID is required")
	}
	db := GetDBInstance(GetDBConfig())
	tx, err := db.Begin()
	if err != nil {
		return err
	}

	query := ""
	curRow, err := db.Query("Select tag_id from item_tags where item_id = ?", itemID)
	if err != nil {
		tx.Rollback()
		return err
	}
	defer curRow.Close()

	currentTags := make(map[string]bool)
	for curRow.Next() {
		var tagID string

		if err := curRow.Scan(&tagID); err != nil {
			tx.Rollback()
			return err
		}
		currentTags[tagID] = true
		if err := curRow.Err(); err != nil {
			tx.Rollback()
			return err
		}
	}

	desiredTags := make(map[string]bool)
	for _, tagID := range tagIDs {
		desiredTags[tagID] = true
	}

	toAdd := []string{}
	toRemove := []string{}

	for key, _ := range desiredTags {
		if !currentTags[key] {
			toAdd = append(toAdd, key)
		}
	}

	for key, _ := range currentTags {
		if !desiredTags[key] {
			toRemove = append(toRemove, key)
		}
	}
	fmt.Println("toAdd", toAdd)
	fmt.Println("toRemove", toRemove)
	fmt.Println("currentTags", currentTags)
	fmt.Println("desiredTags", desiredTags)
	fmt.Println("tagIDs", tagIDs)

	if len(toAdd) == 0 && len(toRemove) == 0 {
		if err := tx.Commit(); err != nil {
			return err
		}
		return nil
	}

	args := make([]interface{}, 0, len(toRemove)+1)
	if len(toRemove) > 0 {
		placeholders := ""
		for i := range toRemove {
			if i > 0 {
				placeholders += ", "
			}
			placeholders += "?"
			args = append(args, toRemove[i])
		}

		query = "DELETE FROM item_tags WHERE item_id = ? AND tag_id IN (" + placeholders + ")"

		if _, err := tx.Exec(query, args...); err != nil {
			fmt.Printf("Error deleting tags: %v\n", err)
			tx.Rollback()
			return err
		}
	}
	if len(toAdd) > 0 {
		stmt, err := tx.Prepare("INSERT INTO item_tags (item_id, tag_id, created_at) VALUES (?, ?, ?)")
		if err != nil {
			tx.Rollback()
			return err
		}
		defer stmt.Close()
		now := time.Now()
		for _, tagId := range toAdd {
			if _, err := stmt.Exec(itemID, tagId, now); err != nil {
				fmt.Printf("Error adding tags: %v\n", err)
				tx.Rollback()
				return err
			}
		}
		if err := stmt.Close(); err != nil {
			fmt.Printf("Error closing statement: %v\n", err)
			tx.Rollback()
			return err
		}
	}
	if err := tx.Commit(); err != nil {
		fmt.Printf("Error committing transaction: %v\n", err)
		return err
	}
	return nil
}

// CreateTag creates a new tag in the database
func CreateTag(tag Tag) (Tag, error) {
	fmt.Println("---CREATETAG---", tag)
	db := GetDBInstance(GetDBConfig())

	// Generate a unique ID if not provided
	if tag.ID == "" {
		tag.ID = fmt.Sprintf("tag_%d", time.Now().UnixNano())
	}

	// Set creation timestamp if not already set
	// if tag.CreatedAt.IsZero() {
	// 	tag.CreatedAt = time.Now()
	// }

	query := "INSERT INTO tags (id, name, category, created_at) VALUES (?, ?, ?, ?)"
	stmt, err := db.Prepare(query)
	if err != nil {
		return Tag{}, err
	}
	defer stmt.Close()

	_, err = stmt.Exec(
		tag.ID,
		tag.TagName,
		// tag.Category,
		// tag.CreatedAt,
	)

	if err != nil {
		return Tag{}, err
	}

	return tag, nil
}

// GetPopularTags retrieves the most frequently used tags
func GetPopularTags(limit int) ([]Tag, error) {
	fmt.Println("---GETPOPULARTAGS---", limit)
	db := GetDBInstance(GetDBConfig())
	var tags []Tag

	if limit <= 0 {
		limit = 10 // Default limit
	}

	query := `
	SELECT t.id, t.name, IFNULL(t.category, ''), t.created_at, COUNT(it.item_id) as usage_count
	FROM tags t
	JOIN item_tags it ON t.id = it.tag_id
	GROUP BY t.id
	ORDER BY usage_count DESC
	LIMIT ?`

	rows, err := db.Query(query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var tag Tag
		var usageCount int
		err := rows.Scan(
			&tag.ID,
			&tag.TagName,
			// &tag.Category,
			// &tag.CreatedAt,
			&usageCount,
		)
		if err != nil {
			return nil, err
		}
		tags = append(tags, tag)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return tags, nil
}

// SearchTagsByName searches for tags by name with a LIKE query
func SearchTagsByName(name string) ([]Tag, error) {
	fmt.Println("---SEARCHTAGSBYNAME---", name)
	db := GetDBInstance(GetDBConfig())
	var tags []Tag

	// Add % wildcards for LIKE query
	searchValue := "%" + name + "%"

	query := "SELECT id, name, IFNULL(category, ''), created_at FROM tags WHERE name LIKE ?"
	rows, err := db.Query(query, searchValue)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var tag Tag
		err := rows.Scan(
			&tag.ID,
			&tag.TagName,
			// &tag.Category,
			// &tag.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		tags = append(tags, tag)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return tags, nil
}

// GetRecommendedItems retrieves items recommended based on the given tags with pagination
func GetRecommendedItems(tagIDs []string, limit int, page int) ([]Item, int, error) {
	fmt.Println("---GETRECOMMENDEDITEMS---", tagIDs, limit, page)
	if len(tagIDs) == 0 {
		return nil, 0, fmt.Errorf("no tags provided")
	}

	if limit <= 0 {
		limit = 10 // Default limit
	}

	if page <= 0 {
		page = 1 // Default to first page
	}

	// Calculate offset based on page number
	offset := (page - 1) * limit

	db := GetDBInstance(GetDBConfig())
	var items []Item

	// Create placeholders for SQL query
	placeholders := ""
	countArgs := make([]interface{}, len(tagIDs))
	args := make([]interface{}, len(tagIDs)+2) // +2 for limit and offset

	for i, tagID := range tagIDs {
		if i > 0 {
			placeholders += ", "
		}
		placeholders += "?"
		countArgs[i] = tagID
		args[i] = tagID
	}
	args[len(tagIDs)] = limit
	args[len(tagIDs)+1] = offset

	// First, get the total count for pagination info
	countQuery := `
	SELECT COUNT(DISTINCT i.item_id)
	FROM items i
	JOIN item_tags it ON i.item_id = it.item_id
	WHERE it.tag_id IN (` + placeholders + `)`

	var totalCount int
	err := db.QueryRow(countQuery, countArgs...).Scan(&totalCount)
	if err != nil {
		return nil, 0, err
	}

	// Query to get recommended items based on tags with pagination
	// This query counts how many of the requested tags each item has
	// Then sorts by the tag match count (most matching tags first)
	query := `
	SELECT i.item_id, IFNULL(i.code, ''), IFNULL(i.barcode, ''), IFNULL(i.box_barcode, ''), IFNULL(i.price, 0), IFNULL(i.box_price, 0), IFNULL(i.name, ''), 
	IFNULL(i.type, ''), IFNULL(i.available_for_order, 0), IFNULL(i.image_path, ''), i.created_at,
	COUNT(it.tag_id) as tag_match_count
	FROM items i
	JOIN item_tags it ON i.item_id = it.item_id
	WHERE it.tag_id IN (` + placeholders + `)
	GROUP BY i.item_id
	ORDER BY tag_match_count DESC, i.created_at DESC
	LIMIT ? OFFSET ?`

	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	for rows.Next() {
		var item Item
		var tagMatchCount int
		err := rows.Scan(
			&item.ID,
			&item.Code,
			&item.BarCode,
			&item.BoxBarcode,
			&item.Name,
			&item.Type,
			&item.AvailableForOrder,
			&item.ImagePath,
			&item.CreatedAt,
			&tagMatchCount,
		)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, item)
	}

	if err = rows.Err(); err != nil {
		return nil, 0, err
	}

	// Fetch tag information for each item
	itemMap := make(map[string]*Item)
	var itemIDs []string

	for i := range items {
		itemMap[items[i].ID] = &items[i]
		itemIDs = append(itemIDs, items[i].ID)
	}

	if len(itemIDs) > 0 {
		// Build item placeholders
		itemPlaceholders := ""
		itemArgs := make([]interface{}, len(itemIDs))

		for i, itemID := range itemIDs {
			if i > 0 {
				itemPlaceholders += ", "
			}
			itemPlaceholders += "?"
			itemArgs[i] = itemID
		}

		// Query to get tags for these items
		tagQuery := `
		SELECT it.item_id, t.id, t.name
		FROM item_tags it
		JOIN tags t ON it.tag_id = t.id
		WHERE it.item_id IN (` + itemPlaceholders + `)`

		tagRows, err := db.Query(tagQuery, itemArgs...)
		if err != nil {
			fmt.Printf("Error fetching tags: %v\n", err)
		} else {
			defer tagRows.Close()

			for tagRows.Next() {
				var itemID string
				var tag Tag

				err := tagRows.Scan(
					&itemID,
					&tag.ID,
					&tag.TagName,
				)
				if err != nil {
					fmt.Printf("Error scanning tag: %v\n", err)
					continue
				}

				// Add tag to the appropriate item
				if item, exists := itemMap[itemID]; exists {
					item.Tag = append(item.Tag, tag)
				}
			}
		}
	}

	// Fetch stock information for each item
	for i := range items {
		stocks, err := GetStocksByItemId(items[i].ID)
		if err != nil {
			// Continue with empty stock if there's an error
			items[i].Stock = []Stock{}
		} else {
			items[i].Stock = stocks
		}
	}

	return items, totalCount, nil
}
