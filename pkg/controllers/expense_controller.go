package controllers

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/ishika-rg/expenseTrackerBackend/pkg/config"
	"github.com/ishika-rg/expenseTrackerBackend/pkg/models"
	"github.com/ishika-rg/expenseTrackerBackend/pkg/utils"

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

	// Send email notification

	// Channel to collect errors from Go routines
	go func() {
		errorChannel := make(chan error, len(expense_body.MemberIDs))

		subject := "New Expense Added"

		for _, each_member := range expense_body.MemberIDs {
			each_member := each_member
			var user models.User
			err := config.DB.First(&user, each_member).Error
			if err != nil {
				return
			}

			email_body := fmt.Sprintf(`
				<p>Hello <strong>%s<strong>,</p>
				<p><strong> %s </strong> has added an expense of <strong>$ %.2f </strong> in the group <strong>%s </strong> .</p>
				<p>If you have any questions or need assistance, feel free to contact our support team.</p>
				<p>Best Regards,</p>
				<p>Monefy Team</p>
			`, getUserNameByID(each_member), new_expense.Expense_paid_by.Name, new_expense.Amount, new_expense.Expense_group.Group_name)

			if err := utils.SendEmail(user.Email, subject, email_body); err != nil {
				errorChannel <- err
			} else {
				errorChannel <- nil
			}

		}
		// Collect errors from Go routines
		for range expense_body.MemberIDs {
			if err := <-errorChannel; err != nil {
				fmt.Println("Failed to send email to some members:", err)
				// Handle email sending errors if needed
			}
		}
	}()

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

	err := config.DB.
		Preload("Expense_paid_by").
		Where("group_id=?", group_id).
		Order("created_at desc").
		Find(&expenses).
		Error

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

	err = config.DB.
		Where("group_id = ? AND (creditor_id = ? OR debtor_id = ?) AND settled=? ", group_id, userid, userid, false).
		Order("created_at desc").
		Find(&transactions).
		Error

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

	err := config.DB.Preload("Expense_details").
		Preload("Expense_details.Expense_paid_by").
		Where("group_id = ? AND settled = ? AND (creditor_id = ? OR debtor_id = ?)", group_id, false, user_id, user_id).
		Order("created_at desc").
		Find(&unsettledTransaction).Error
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
	// After the settlement is created, loading the associated transaction details
	if err := tx.Preload("Transaction").Preload("Transaction.Expense_details").First(&settlement, settlement.ID).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"data":    gin.H{},
			"error":   true,
			"status":  500,
			"message": "Failed to load transaction details",
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
	fmt.Println("process to send mails has started....")

	// fmt.Println("mail ids are : ")
	// fmt.Println(settlement.Transaction.Creditor_id)
	var creditor models.User
	errr := config.DB.First(&creditor, settlement.Transaction.Creditor_id).Error
	if errr != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error here": errr.Error(),
		},
		)
	}
	fmt.Println(creditor.Email)
	var debtor models.User
	er := config.DB.First(&debtor, settlement.Transaction.Debtor_id).Error
	if er != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error here": er.Error(),
		},
		)
	}
	fmt.Println(debtor.Email)

	// Send email notification after successfull settlement
	subject := "Expense Settled!"
	go func() {

		email_body_creditor := fmt.Sprintf(`
		
				<p>Hello <strong>%s<strong>,</p>
				<p><strong> %s </strong> has made a settlement of <strong>$ %.2f </strong> for the expense <strong>%s </strong> .</p>

				<p>If you have any questions or need assistance, feel free to contact our support team.</p>
				<p>Best Regards,</p>
				<p>Monefy Team</p>
			`, creditor.Name, getUserNameByID(settlement.SettledBy), settlement.SettledAmount, settlement.Transaction.Expense_details.Expense_name)

		if err := utils.SendEmail(creditor.Email, subject, email_body_creditor); err != nil {
			c.JSON(http.StatusInternalServerError,
				gin.H{
					"error":   "settlement done but failed to send email",
					"message": "Settlement done  but failed to send email",
					"status":  500,
				})
			fmt.Println(err)
			return
		}
		email_body_debitor := fmt.Sprintf(`
		
				<p>Hello <strong>%s<strong>,</p>
				<p><strong> %s </strong> has made a settlement of <strong>$ %.2f </strong> for the expense <strong>%s </strong> .</p>

				<p>If you have any questions or need assistance, feel free to contact our support team.</p>
				<p>Best Regards,</p>
				<p>Monefy Team</p>
			`, debtor.Name, getUserNameByID(settlement.SettledBy), settlement.SettledAmount, settlement.Transaction.Expense_details.Expense_name)

		if err := utils.SendEmail(debtor.Email, subject, email_body_debitor); err != nil {
			c.JSON(http.StatusInternalServerError,
				gin.H{
					"error":   "Settlement done but failed to send email",
					"message": "Settlement done but failed to send email",
					"status":  500,
				})
			fmt.Println(err)
			return
		}
	}()

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

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"error":   false,
		"message": "Settlements retrieved successfully",
		"status":  200,
		"data": gin.H{
			"settlements": settlements,
			"info":        response,
		},
	})
}

func GetUserTransactionsWithMembers(c *gin.Context) {
	userID := c.Param("user_id")
	// Convert userID to an integer
	loggedInUserID, err := strconv.Atoi(userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   true,
			"message": "Invalid User ID",
			"status":  400,
		})
		return
	}
	type TransactionSummary struct {
		ID             int     `json:"id"`
		GroupID        int     `json:"group_id"`
		GroupDesc      string  `json:"group_desc"`
		GroupName      string  `json:"group_name"`
		GroupAdminID   int     `json:"group_admin_id"`
		GroupAdminName string  `json:"group_admin_name"`
		CreditorName   string  `json:"creditor_name"`
		DebtorName     string  `json:"debtor_name"`
		ExpenseAmount  float64 `json:"expense_amount"`
		CreditbyId     int     `json:"creditby_id"`

		// Add more fields as needed

	}
	type MemberTransactionSummary struct {
		MemberID       int                  `json:"member_id"`
		MemberName     string               `json:"member_name"`
		OverallAmount  float64              `json:"overall_amount"`
		Status         string               `json:"status"`
		PendingDetails []TransactionSummary `json:"pending_details"`
	}

	// var transactions []models.Transactions

	var transactions []models.Transactions

	// Fetch all transactions involving the logged-in user
	if err := config.DB.
		Where("(creditor_id = ? OR debtor_id = ?) AND settled IS NOT NULL", loggedInUserID, loggedInUserID).
		Preload("Expense_details").
		Preload("GroupDetails").
		Preload("GroupDetails.Admin").
		Find(&transactions).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch transactions", "details": err.Error()})
		return
	}

	//  Group by each member
	memberSummary := make(map[int]*MemberTransactionSummary)
	for _, transaction := range transactions {
		var memberID int
		var isCreditor bool

		// Determine the member and the role of the logged-in user in this transaction
		if transaction.Creditor_id == loggedInUserID {
			memberID = transaction.Debtor_id
			isCreditor = true
		} else {
			memberID = transaction.Creditor_id
			isCreditor = false
		}

		// Initialize the member summary if not already done
		if _, exists := memberSummary[memberID]; !exists {
			memberSummary[memberID] = &MemberTransactionSummary{
				MemberID:       memberID,
				MemberName:     getUserNameByID(memberID),
				OverallAmount:  0,
				Status:         "All completed", // Default status
				PendingDetails: []TransactionSummary{},
			}
		}

		// Calculating overall amount and check for unsettled transactions
		if !transaction.Settled {
			memberSummary[memberID].Status = "Not completed"
			memberSummary[memberID].PendingDetails = append(memberSummary[memberID].PendingDetails,
				TransactionSummary{
					ID:             int(transaction.ID),
					GroupID:        int(transaction.Group_id),
					GroupName:      transaction.GroupDetails.Group_name,
					GroupDesc:      transaction.GroupDetails.Description,
					GroupAdminID:   transaction.GroupDetails.Group_admin_id,
					GroupAdminName: transaction.GroupDetails.Admin.Name,
					CreditorName:   getUserNameByID(transaction.Creditor_id),
					DebtorName:     getUserNameByID(transaction.Debtor_id),
					ExpenseAmount:  transaction.Amount,
					CreditbyId:     int(transaction.Creditor_id),
				},
			)
		}

		if isCreditor && !transaction.Settled {
			memberSummary[memberID].OverallAmount += transaction.Amount
		} else if !isCreditor && !transaction.Settled {
			memberSummary[memberID].OverallAmount -= transaction.Amount
		}
	}

	// forming the result set for response
	var result []MemberTransactionSummary
	for _, summary := range memberSummary {
		result = append(result, *summary)
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    result,
		"message": "Member wise transactions fetched successfully",
		"status":  200,
		"error":   false,
	})

}
