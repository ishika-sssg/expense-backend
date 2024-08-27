package models

import (
	"time"

	"github.com/jinzhu/gorm"
)

type Expense struct {
	gorm.Model

	Expense_name       string  `gorm:"column:expense_name" json:"expense_name"`
	Expense_desc       string  `gorm:"column:expense_desc" json:"expense_desc"`
	Amount             float64 `gorm:"column:amount" json:"amount"`
	User_id            int     `gorm:"not null" json:"user_id"`                      // Foreign key referencing User.ID
	Expense_created_by User    `gorm:"foreignKey:user_id" json:"expense_created_by"` // Association
	Group_id           int     `gorm:"not null" json:"group_id"`
	Expense_group      Group   `gorm:"foreignKey:group_id" json:"expense_group"`
	Paid_by            int     `gorm:"not null" json:"paid_by"`
	Expense_paid_by    User    `gorm:"foreignKey:Paid_by" json:"expense_paid_by"` // Association

}

type ExpenseShare struct {
	gorm.Model

	Expense_id     int     `gorm:"not null" json:"expense_id"`
	Member_id      int     `gorm:"not null" json:"member_id"`
	Amount_share   float64 `gorm:"column:amount_share" json:"amount_share"`
	Expense_detail Expense `gorm:"foreignKey:expense_id" json:"expense_detail"`
	Member_detail  User    `gorm:"foreignKey:member_id" json:"member_detail"`
}

type Transactions struct {
	gorm.Model

	Creditor_id     int     `gorm:"not null" json:"creditor_id"` //who paid for the expense
	Debtor_id       int     `gorm:"not null" json:"debtor_id"`   //who owes money, who will give money to other
	Amount          float64 `gorm:"column:amount" json:"amount"`
	Expense_id      int     `gorm:"not null" json:"expense_id"`
	Group_id        int     `gorm:"not null" json:"group_id"`
	Expense_details Expense `gorm:"foreignKey:expense_id" json:"expense_details"`
	Settled         bool    `gorm:"column:settled" json:"settled"` // New field to indicate settlement status

	// added group_details
	GroupDetails Group `gorm:"foreignKey:group_id" json:"group_details"`
}

type Settlement struct {
	gorm.Model

	TransactionID int          `gorm:"not null" json:"transaction_id"` // Reference to the transaction being settled
	SettledAmount float64      `gorm:"column:settled_amount" json:"settled_amount"`
	SettledBy     int          `gorm:"not null" json:"settled_by"`                  // User ID who performed the settlement
	Transaction   Transactions `gorm:"foreignKey:TransactionID" json:"transaction"` // Association
	SettledAt     time.Time    `gorm:"column:settled_at" json:"settled_at"`
}

type ExpenseInfoResponse struct {
	UserShare float64 `json:"user_share"` // Amount the user needs to pay
	UserOwed  float64 `json:"user_owed"`  //Amount user will get
}

type OneToOneDetail struct {
	UserID     int     `json:"user_id"`
	UserName   string  `json:"user_name"`
	Amount     float64 `json:"amount"`
	IsCreditor bool    `json:"is_creditor"`
}

type SummaryResponse struct {
	OneToOneDetails []OneToOneDetail `json:"one_to_one_details"`
	TotalLent       float64          `json:"total_lent"`
	TotalBorrowed   float64          `json:"total_borrowed"`
	NetBalance      float64          `json:"net_balance"` // Positive if the user is owed, negative if the user owes
}

type SettlementResponse struct {
	SettledAtDate string  `json:"settled_at_date"` // Formatted time string
	CreditorName  string  `json:"creditor_name"`
	DebtorName    string  `json:"debtor_name"`
	Amount        float64 `json:"amount"`
	SettlerName   string  `json:"settler_name"`
	Cred_id       int     `json:"cred_id"`
	Deb_id        int     `json:"deb_id"`
	Sett_id       int     `json:"sett_id"`
}
