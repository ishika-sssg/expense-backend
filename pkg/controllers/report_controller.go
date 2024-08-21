package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/ishika-rg/expenseTrackerBackend/pkg/config"
	"github.com/ishika-rg/expenseTrackerBackend/pkg/models"

	"encoding/csv"
	"strconv"
)

func GenerateCsvSettlements(c *gin.Context) {
	user_id := c.Param("user_id")

	// retrieving settlement of

	var settlements []models.Settlement

	// Query settlements where the user is either the one who settled or involved in the transaction
	err := config.DB.
		Preload("Transaction").
		Preload("Transaction.Expense_details").
		Where("settled_by = ? OR transaction_id IN (SELECT id FROM transactions WHERE creditor_id = ? OR debtor_id = ?)", user_id, user_id, user_id).
		Order("created_at desc").
		Find(&settlements).Error

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   true,
			"message": "Error retrieving settlements",
			"status":  500,
			"data":    err.Error(),
		})
		return
	}
	var response []models.SettlementResponse
	for _, settlement := range settlements {
		response = append(response, models.SettlementResponse{

			SettledAtDate: settlement.SettledAt.Format("02 Jan 2006"), // Format the time as a readable string
			CreditorName:  getUserNameByID(settlement.Transaction.Creditor_id),
			DebtorName:    getUserNameByID(settlement.Transaction.Debtor_id),
			Amount:        settlement.Transaction.Amount,
			SettlerName:   getUserNameByID(settlement.SettledBy),
			Cred_id:       settlement.Transaction.Creditor_id,
			Deb_id:        settlement.Transaction.Debtor_id,
			Sett_id:       settlement.SettledBy,
		})
	}

	// file, err := os.Create("Settlement_records.csv")
	// // defer file.Close()
	// if err != nil {
	// 	log.Fatalln("failed to open file", err)
	// }

	// Create a CSV writer
	c.Writer.Header().Set("Content-Type", "text/csv")
	c.Writer.Header().Set("Content-Disposition", "attachment;filename=settlements.csv")
	csvWriter := csv.NewWriter(c.Writer)
	defer csvWriter.Flush()

	// Writeing CSV headers
	headers := []string{"ID", "Creditor", "Debtor", "Settled Amount", "Settled By", "Settled At"}
	if err := csvWriter.Write(headers); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   true,
			"message": "Error writing CSV headers",
			"status":  500,
			"data":    err.Error(),
		})
		return
	}

	// Write data rows
	for i, settlement := range response {
		row := []string{
			strconv.Itoa(i + 1),
			// strconv.Itoa(settlement.CreditorName),
			settlement.CreditorName,
			settlement.DebtorName,

			strconv.FormatFloat(settlement.Amount, 'f', 2, 64),
			settlement.SettlerName,
			settlement.SettledAtDate,
		}
		if err := csvWriter.Write(row); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"error":   true,
				"message": "Error writing CSV row",
				"status":  500,
				"data":    err.Error(),
			})
			return
		}
	}

}
