package controllers

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/ishika-rg/expenseTrackerBackend/pkg/config"
	"github.com/ishika-rg/expenseTrackerBackend/pkg/models"

	"strconv"
	"time"
)

func CreateExpense(c *gin.Context) {
	// get group and its details from body
	user_id, exists := c.Get("user_id")
	fmt.Println(user_id, exists)

	var expense_body struct {
		Expense_name string  `json:"expense_name" binding:"required"`
		Expense_desc string  `json:"expense_desc"`
		Amount       float64 `json:"amount"`
		Group_id     int     `json:"group_id"`
		Paid_by      int     `json:"paid_by"`
		MemberIDs    []int   `json:"member_ids"`
	}

	if err := c.Bind(&expense_body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   true,
			"data":    gin.H{},
			"message": "Failed to read body " + err.Error(),
			"status":  400,
		})
		return
	}

	// split amount between selected membersof group :
	splitAmount := expense_body.Amount / float64(len(expense_body.MemberIDs))

	// Start a new database transaction
	tx := config.DB.Begin()
	// Check for any errors in starting the transaction
	if tx.Error != nil {
		c.JSON(http.StatusInternalServerError,
			gin.H{
				"success": false,
				"error":   true,
				"data":    gin.H{},
				"message": "Failed to read body",
				"status":  400,
			})
		return
	}

	new_expense := models.Expense{
		Expense_name: expense_body.Expense_name,
		Expense_desc: expense_body.Expense_desc,
		Amount:       expense_body.Amount,
		// User_id:      expense_body.User_id,
		Group_id: expense_body.Group_id,
		Paid_by:  expense_body.Paid_by,
	}

	if err := tx.Create(&new_expense).Error; err != nil {
		tx.Rollback() // Rollback the transaction in case of error
		c.JSON(http.StatusInternalServerError,
			gin.H{
				"success": false,
				"error":   true,
				"data":    gin.H{},
				"message": "Error :Failed to create expense",
				"status":  400,
			})
		return
	}

	if err := tx.Preload("Expense_created_by").Preload("Expense_group").Preload("Expense_group.Admin").Preload("Expense_paid_by").First(&new_expense, new_expense.ID).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError,
			gin.H{
				"success": false,
				"error":   true,
				"data":    gin.H{},
				"message": "Error : Failed to load details",
				"status":  400,
			})
		return
	}

	// Create member share for each member
	fmt.Println("here my another tx start")
	for _, each_member_id := range expense_body.MemberIDs {
		each_share := models.ExpenseShare{
			Expense_id:   int(new_expense.ID),
			Member_id:    int(each_member_id),
			Amount_share: splitAmount,
		}

		if err := tx.Create(&each_share).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error from member share": err.Error()})
			return
		}
		// fmt.Printf("Created transaction: %+v", each_share)
		fmt.Println("each transaction here")

	}

	// create one to one trasaction, each time an expense is created :

	for _, single_member_id := range expense_body.MemberIDs {
		if single_member_id != int(expense_body.Paid_by) {
			transaction := models.Transactions{
				Creditor_id: expense_body.Paid_by,
				Debtor_id:   int(single_member_id),
				Amount:      splitAmount,
				Expense_id:  int(new_expense.ID),
				Group_id:    expense_body.Group_id,
				Settled:     false,
			}

			if err := tx.Create(&transaction).Error; err != nil {
				tx.Rollback()
				c.JSON(http.StatusInternalServerError, gin.H{"error from creating transaction table": err.Error()})
				return
			}
			// fmt.Printf("Created transaction: %+v", transaction)
			fmt.Println("each ONE TO ONE transaction here")
		}
	}

	// Commit the transaction
	if err := tx.Commit().Error; err != nil {
		tx.Rollback() // Rollback if commit fails
		c.JSON(http.StatusInternalServerError,
			gin.H{
				"success": false,
				"error":   true,
				"data":    gin.H{},
				"message": "Error : Failed to commit transaction",
				"status":  400,
			})
		return
	}

	// once transaction os complete, return result
	c.JSON(http.StatusOK,
		gin.H{
			"success": true,
			"error":   false,
			"data": gin.H{
				"details": new_expense,
			},
			"message": "Expense created successfully",
			"status":  200,
		},
	)

}

func GetAllExpensesByGroupId(c *gin.Context) {

	group_id := c.Param("group_id")
	user_id := c.Param("user_id")
	var expenses []models.Expense

	// Convert the ID from string to integer
	userid, error := strconv.Atoi(user_id)
	if error != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

<<<<<<< HEAD
	err := config.DB.
		Preload("Expense_paid_by").
		Where("group_id=?", group_id).
		Order("created_at desc").
		Find(&expenses).
		Error
=======
	err := config.DB.Preload("Expense_paid_by").Where("group_id=?", group_id).Find(&expenses).Error
>>>>>>> 16a4be235ba0645c7b0722b6fa6a7290944014be

	fmt.Println(err)
	// fmt.Println(group)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   true,
			"message": "No expenses found",
			"status":  400,
			"data":    err,
		})
		return
	}

	// for dynamic response:
	var response []gin.H

	// for individual calculations :
	for _, expenseShare := range expenses {
		var usershare, userOwed float64

		var transactions []models.Transactions
		err = config.DB.Where("expense_id=?", expenseShare.ID).Find(&transactions).Error
		if err == nil {
			for _, ele := range transactions {
				if ele.Debtor_id == userid {
					usershare += ele.Amount
				}
				if ele.Creditor_id == userid {
					userOwed += ele.Amount
				}
			}
		}

		// Create a response object for the current expense
		expenseData := gin.H{
			"expense_id":   expenseShare.ID,
			"expense_name": expenseShare.Expense_name,
			"expense_desc": expenseShare.Expense_desc,
			"amount":       expenseShare.Amount,
			"created_by":   expenseShare.Expense_created_by,
			"paid_by":      expenseShare.Expense_paid_by,
			"group_id":     expenseShare.Group_id,
			"user_share":   usershare, // Adding userShare to response
			"user_owed":    userOwed,  // Adding userOwed to response
		}

		// Add the expense data to the response slice
		response = append(response, expenseData)

	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"error":   false,
		"message": "All Expenses",
		"status":  200,
		"data": gin.H{
			"alldata":   expenses,
			"total":     len(expenses),
			"shareinfo": response,
		},
	})

}

func GetAllTransactions(c *gin.Context) {
	group_id := c.Param("group_id")
	user_id := c.Param("user_id")

	var transactions []models.Transactions
	// Convert the user ID from string to integer
	userid, err := strconv.Atoi(user_id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	// making changes here .......
	// Fetch all transactions for the group where the user is either a debtor or creditor

<<<<<<< HEAD
	err = config.DB.
		Where("group_id = ? AND (creditor_id = ? OR debtor_id = ?) AND settled=? ", group_id, userid, userid, false).
		Order("created_at desc").
		Find(&transactions).
		Error

=======
	err = config.DB.Where("group_id = ? AND (creditor_id = ? OR debtor_id = ?) AND settled=? ", group_id, userid, userid, false).Find(&transactions).Error
>>>>>>> 16a4be235ba0645c7b0722b6fa6a7290944014be
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   true,
			"message": "No transactions found",
			"status":  400,
			"data":    err,
		})
		return
	}

	// Maps to track total lent and borrowed per user
	userLentMap := make(map[int]float64)
	userBorrowedMap := make(map[int]float64)

	// Variables to store overall totals
	var totalLent, totalBorrowed float64

	// Loop through the transactions and calculate totals
	for _, transaction := range transactions {
		if transaction.Creditor_id == userid {
			// User lent money to someone
			totalLent += transaction.Amount
			userLentMap[transaction.Debtor_id] += transaction.Amount
		} else if transaction.Debtor_id == userid {
			// User borrowed money from someone
			totalBorrowed += transaction.Amount
			userBorrowedMap[transaction.Creditor_id] += transaction.Amount
		}
	}

	// Prepare the detailed one-to-one list
	var oneToOneDetails []models.OneToOneDetail
	for userID, amount := range userLentMap {
		oneToOneDetails = append(oneToOneDetails, models.OneToOneDetail{
			UserID:     userID,
			UserName:   getUserNameByID(userID),
			Amount:     amount,
			IsCreditor: true,
		})
	}
	for userID, amount := range userBorrowedMap {
		oneToOneDetails = append(oneToOneDetails, models.OneToOneDetail{
			UserID:     userID,
			UserName:   getUserNameByID(userID),
			Amount:     amount,
			IsCreditor: false,
		})
	}

	// Calculate net balance
	netBalance := totalLent - totalBorrowed

	// Prepare the summary response
	summary := models.SummaryResponse{
		OneToOneDetails: oneToOneDetails,
		TotalLent:       totalLent,
		TotalBorrowed:   totalBorrowed,
		NetBalance:      netBalance,
	}

	// Prepare the overall details
	overallDetails := make(map[int]float64)
	for _, detail := range oneToOneDetails {
		if detail.IsCreditor {
			overallDetails[detail.UserID] += detail.Amount
		} else {
			overallDetails[detail.UserID] -= detail.Amount
		}
	}

	// Convert overallDetails to the desired format
	var overallDetailsList []gin.H
	for userID, amount := range overallDetails {
		status := "needs to receive"
		if amount < 0 {
			status = "needs to pay"
			amount = -amount // Convert to positive for display
		}
		overallDetailsList = append(overallDetailsList, gin.H{
			"user_id":    userID,
			"user_name":  getUserNameByID(userID),
			"net_amount": amount,
			"status":     status,
		})
	}

	// Return the summary response as JSON
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"error":   false,
		"message": "Expense Summary",
		"status":  200,
		"data": gin.H{
			"summary":         summary,
			"overall_details": overallDetailsList,
		},
	})

}

func getUserNameByID(userID int) string {
	var user models.User
	err := config.DB.First(&user, userID).Error
	if err != nil {
		return "Unknown User"
	}
	return user.Name
}

func GetAllUnsettledTransByGroupId(c *gin.Context) {
	group_id := c.Param("group_id")
	user_id := c.Param("user_id")
	var unsettledTransaction []models.Transactions

<<<<<<< HEAD
	err := config.DB.Preload("Expense_details").
		Preload("Expense_details.Expense_paid_by").
		Where("group_id = ? AND settled = ? AND (creditor_id = ? OR debtor_id = ?)", group_id, false, user_id, user_id).
		Order("created_at desc").
		Find(&unsettledTransaction).Error
=======
	err := config.DB.Preload("Expense_details").Preload("Expense_details.Expense_paid_by").Where("group_id = ? AND settled = ? AND (creditor_id = ? OR debtor_id = ?)", group_id, false, user_id, user_id).Find(&unsettledTransaction).Error
>>>>>>> 16a4be235ba0645c7b0722b6fa6a7290944014be
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   true,
			"message": "Error retrieving unsettled transactions",
			"status":  500,
			"data":    err.Error(),
		})
		return
	}
	type TransactionResponse struct {
		models.Transactions
		CreditorName string  `json:"creditor_name"`
		DebtorName   string  `json:"debtor_name"`
		SettleAmount float64 `json:"settle_amount"`
	}
	var response []TransactionResponse

	// Adding creditor and debtor names to each transaction
	for _, transaction := range unsettledTransaction {
		creditorName := getUserNameByID(transaction.Creditor_id)
		debtorName := getUserNameByID(transaction.Debtor_id)
		SettleAmount := transaction.Amount

		// Appending the transaction with the additional name fields
		response = append(response, TransactionResponse{
			Transactions: transaction,
			CreditorName: creditorName,
			DebtorName:   debtorName,
			SettleAmount: SettleAmount,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"error":   false,
		"message": "Unsettled transactions retrieved successfully",
		"status":  200,
		"data": gin.H{
			"transactions": response,
		},
	})

}

func SettleTransactions(c *gin.Context) {
	transactionID := c.Param("transaction_id")
	userID := c.Param("user_id")

	// Parsing IDs
	txnID, err := strconv.Atoi(transactionID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid transaction ID"})
		return
	}

	uid, err := strconv.Atoi(userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	// Starting a new transaction
	tx := config.DB.Begin()
	if tx.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{

			"message": "Failed to start transaction",
			"success": false,
			"data":    gin.H{},
			"error":   true,
			"status":  500,
		})
		return
	}

	//  Locking the transaction record for update
	var transaction models.Transactions
	if err := tx.Set("gorm:query_option", "FOR UPDATE").Where("id = ? AND settled = ?", txnID, false).First(&transaction).Error; err != nil {
		tx.Rollback() // Rollback the transaction in case of error
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"data":    gin.H{},
			"error":   true,
			"status":  404,
			"message": "Transaction not found or already settled",
		})
		return
	}

	//  Updating the transaction's settled status
	transaction.Settled = true
	if err := tx.Save(&transaction).Error; err != nil {
		// Rollback the transaction in case of error
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"data":    gin.H{},
			"error":   true,
			"status":  500,
			"message": "Failed to update transaction status"})
		return
	}

	// Creating a settlement record
	settlement := models.Settlement{
		TransactionID: txnID,
		SettledAmount: transaction.Amount,
		SettledBy:     uid,
		SettledAt:     time.Now(),
	}

	if err := tx.Create(&settlement).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"data":    gin.H{},
			"error":   true,
			"status":  500,
			"message": "Failed to create settlement record",
		})
		return
	}

	// Commit the transaction
	if err := tx.Commit().Error; err != nil {
		tx.Rollback() // Rollback if the commit fails
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Failed to commit transaction",
			"success": false,
			"data":    gin.H{},
			"error":   true,
			"status":  500,
		})
		return
	}

	// If everything is successful, return a success response
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Transaction settled successfully",
		"data":    settlement,
		"error":   false,
		"status":  200,
	})

}

func GetAllSettlementRecord(c *gin.Context) {
	user_id := c.Param("user_id")
	var settlements []models.Settlement

	// Query settlements where the user is either the one who settled or involved in the transaction
	err := config.DB.
		Preload("Transaction").
		Preload("Transaction.Expense_details").
		Where("settled_by = ? OR transaction_id IN (SELECT id FROM transactions WHERE creditor_id = ? OR debtor_id = ?)", user_id, user_id, user_id).
<<<<<<< HEAD
		Order("created_at desc").
=======
>>>>>>> 16a4be235ba0645c7b0722b6fa6a7290944014be
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

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"error":   false,
		"message": "Settlements retrieved successfully",
		"status":  200,
		"data": gin.H{
			"settlements": settlements,
		},
	})
}
