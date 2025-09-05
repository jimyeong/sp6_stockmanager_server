package apis

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/jimyeongjung/owlverload_api/models"
)

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
