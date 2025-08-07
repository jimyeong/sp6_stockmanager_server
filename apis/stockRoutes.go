package apis

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/jimyeongjung/owlverload_api/models"
	"github.com/jimyeongjung/owlverload_api/utils"
)

func HandleStockUpdate(w http.ResponseWriter, r *http.Request) {

	// update stock info, (expiry date, location, discount rate)
	// get the stock id from the request body
	// will return the whole stockc info with the given stock id

	var stock models.Stock
	err := json.NewDecoder(r.Body).Decode(&stock)
	if err != nil {
		utils.Error("Error decoding stock update request: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if stock.StockId == "" {
		utils.Error("stock_id is required")
		http.Error(w, "stock_id is required", http.StatusBadRequest)
		return
	}

	db := models.GetDBInstance(models.GetDBConfig())
	query := "UPDATE stocks SET expiry_date = ?, location = ?, discount_rate = ? WHERE stock_id = ?"
	_, err = db.Exec(query, stock.ExpiryDate, stock.Location, stock.DiscountRate, stock.StockId)
	if err != nil {
		fmt.Println("@@@err@@@@@@@@@@@@@@@@@@@@@", stock)
		utils.Error("Error updating stock: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var updatedStock models.Stock
	querySelect := "SELECT stock_id, fkproduct_id, box_number, single_number, bundle_number, expiry_date, location, registering_person, notes, discount_rate, created_at FROM stocks WHERE stock_id = ?"
	err = db.QueryRow(querySelect, stock.StockId).Scan(
		&updatedStock.StockId,
		&updatedStock.ItemId,
		&updatedStock.BoxNumber,
		&updatedStock.SingleNumber,
		&updatedStock.BundleNumber,
		&updatedStock.ExpiryDate,
		&updatedStock.Location,
		&updatedStock.RegisteringPerson,
		&updatedStock.Notes,
		&updatedStock.DiscountRate,
		&updatedStock.CreatedAt,
	)
	if err != nil {
		utils.Error("Error selecting stock: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	utils.Info("Stock updated successfully")
	models.WriteServiceResponse(w, "Stock updated successfully", updatedStock, true, true, http.StatusOK)
}
