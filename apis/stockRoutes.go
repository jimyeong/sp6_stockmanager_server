package apis

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/jimyeongjung/owlverload_api/models"
)

type GetStockByItemIdResponse struct {
	ItemId string         `json:"itemId"`
	Total  int            `json:"total"`
	Stocks []models.Stock `json:"stocks"`
}

func HandleStockUpdate(w http.ResponseWriter, r *http.Request) {

	// update stock info, (expiry date, location, discount rate)
	// get the stock id from the request body
	// will return the whole stockc info with the given stock id

	var stock models.Stock
	err := json.NewDecoder(r.Body).Decode(&stock)
	if err != nil {
		fmt.Println("---Error decoding stock update request: %v---", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if stock.StockId == "" {
		fmt.Println("---stock_id is required---")
		http.Error(w, "stock_id is required", http.StatusBadRequest)
		return
	}

	db := models.GetDBInstance(models.GetDBConfig())
	query := "UPDATE stocks SET expiry_date = ?, location = ?, discount_rate = ? WHERE stock_id = ?"
	_, err = db.Exec(query, stock.ExpiryDate, stock.Location, stock.DiscountRate, stock.StockId)
	if err != nil {
		fmt.Println("---Error updating stock: %v---", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var updatedStock models.Stock
	querySelect := "SELECT stock_id, fkproduct_id, stock_type, box_number, pcs_number, bundle_number, expiry_date, location, registering_person, notes, discount_rate, created_at FROM stocks WHERE stock_id = ?"
	err = db.QueryRow(querySelect, stock.StockId).Scan(
		&updatedStock.StockId,
		&updatedStock.ItemId,
		&updatedStock.StockType,
		&updatedStock.BoxNumber,
		&updatedStock.PCSNumber,
		&updatedStock.BundleNumber,
		&updatedStock.ExpiryDate,
		&updatedStock.Location,
		&updatedStock.RegisteringPerson,
		&updatedStock.Notes,
		&updatedStock.DiscountRate,
		&updatedStock.CreatedAt,
	)
	if err != nil {
		fmt.Println("---Error selecting stock: %v---", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	models.WriteServiceResponse(w, "Stock updated successfully", updatedStock, true, true, http.StatusOK)
}

func HandleExpiryStockCheck(w http.ResponseWriter, r *http.Request) {

	// check if the stock is expired
	// get the stock id from the request body
	// will return the whole stockc info with the given stock id

}

func HandleGetStockByItemId(w http.ResponseWriter, r *http.Request) {
	// get the item id from the query string
	itemId := r.URL.Query().Get("itemId")
	if itemId == "" {
		// keep error style consistent with this file
		http.Error(w, "itemId is required", http.StatusBadRequest)
		return
	}

	db := models.GetDBInstance(models.GetDBConfig())
	if db == nil {
		http.Error(w, "database connection error", http.StatusInternalServerError)
		return
	}

	query := `
		SELECT stock_id, fkproduct_id, stock_type, box_number, pcs_number, bundle_number,
			expiry_date, location, registering_person, IFNULL(notes, ''), IFNULL(discount_rate, 0), created_at
		FROM stocks
		WHERE fkproduct_id = ?
		ORDER BY expiry_date ASC, created_at DESC
	`

	rows, err := db.Query(query, itemId)
	if err != nil {
		fmt.Println("---Error querying stocks by itemId: %v---", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	stocks := make([]models.Stock, 0)
	for rows.Next() {
		var s models.Stock
		err := rows.Scan(
			&s.StockId,
			&s.ItemId,
			&s.StockType,
			&s.BoxNumber,
			&s.PCSNumber,
			&s.BundleNumber,
			&s.ExpiryDate,
			&s.Location,
			&s.RegisteringPerson,
			&s.Notes,
			&s.DiscountRate,
			&s.CreatedAt,
		)
		if err != nil {
			fmt.Println("---Error scanning stock row: %v---", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		stocks = append(stocks, s)
	}
	if err := rows.Err(); err != nil {
		fmt.Println("---Row iteration error: %v---", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := GetStockByItemIdResponse{
		ItemId: itemId,
		Total:  len(stocks),
		Stocks: stocks,
	}
	models.WriteServiceResponse(w, "Stocks retrieved successfully", response, true, true, http.StatusOK)
}
