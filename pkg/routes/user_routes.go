package routes

import (
	// "github.com/gorilla/mux"
	"github.com/gin-gonic/gin"
	"github.com/ishika-rg/expenseTrackerBackend/pkg/controllers"
	"github.com/ishika-rg/expenseTrackerBackend/pkg/middleware"
)

func SetRouter() *gin.Engine {
	r := gin.Default()
	auth := r.Group("/auth")
	{
		auth.POST("/register", controllers.Signup)
		auth.POST("/login", controllers.Login)
		auth.GET("/user/profile", middleware.RequireAuth, controllers.Validate)
	}

	// group route for authenticated user
	group := r.Group("/group")
	{
		group.POST("", middleware.RequireAuth, controllers.CreateGroup)
		group.GET("/:user_email", middleware.RequireAuth, controllers.GetMemberByEmail)
		group.POST("/add_group_member", middleware.RequireAuth, controllers.AddGroupMemberByEmail)
		group.GET("/all/:user_id", middleware.RequireAuth, controllers.GetAllGroupsById)
		group.GET("/get_allmembers/:group_id", middleware.RequireAuth, controllers.GetAllGroupMembersByGroupId)
		group.DELETE("/delete/group_id/:group_id/member_id/:member_id/group_admin_id/:group_admin_id/loggedin_userid/:loggedin_user_id", middleware.RequireAuth, controllers.DeleteMemberFromGroup)
	}

	expense := r.Group("/expense")
	{
		expense.POST("/add", middleware.RequireAuth, controllers.CreateExpense)
		expense.GET("/allexpenses/:group_id/user_id/:user_id", middleware.RequireAuth, controllers.GetAllExpensesByGroupId)
		expense.GET("/expensedata/group_id/:group_id/user_id/:user_id", middleware.RequireAuth, controllers.GetAllTransactions)
		expense.GET("/unsettledtransactions/group_id/:group_id/user_id/:user_id", middleware.RequireAuth, controllers.GetAllUnsettledTransByGroupId)
		expense.POST("/settletransaction/transid/:transaction_id/user_id/:user_id", middleware.RequireAuth, controllers.SettleTransactions)
		expense.GET("/allsettlements/user_id/:user_id", middleware.RequireAuth, controllers.GetAllSettlementRecord)
		expense.GET("/members_expense/user_id/:user_id", middleware.RequireAuth, controllers.GetUserTransactionsWithMembers)
		expense.GET("/csv/settlerecord/user_id/:user_id", middleware.RequireAuth, controllers.GenerateCsvSettlements)
	}

	return r
}
