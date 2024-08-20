package controllers

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/ishika-rg/expenseTrackerBackend/pkg/config"
	"github.com/ishika-rg/expenseTrackerBackend/pkg/models"
	"github.com/jinzhu/gorm"
)

func CreateGroup(c *gin.Context) {

	// get group and its details from body
	user_id, exists := c.Get("user_id")
	fmt.Println(user_id, exists)

	var body struct {
		Group_name     string `json:"group_name"`
		Description    string `json:"description"`
		Group_admin_id int    `json:"group_admin_id"`
		Category       string `json:"category"`
	}

	if c.Bind(&body) != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Failed to read body",
		})
		return
	}

	new_group := models.Group{
		Group_name:     body.Group_name,
		Description:    body.Description,
		Group_admin_id: body.Group_admin_id,
		// Group_admin_id: int(user_id.(uint)),

		Category: body.Category,
	}

	// fmt.Printf("rrr %+v\n ", new_group)

	// get group admin details :
	// Fetch admin details
	var admin models.User
	adminRes := config.DB.First(&admin, new_group.Group_admin_id)
	if adminRes.Error != nil {
		fmt.Println(adminRes.Error)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   true,
			"data":    gin.H{},
			"message": "Error fetching admin details",
			"status":  500,
		})
		return
	}

	res := config.DB.Create(&new_group)
	if res.Error != nil {
		fmt.Println(res.Error)
		c.JSON(http.StatusBadRequest, gin.H{

			"success": false,
			"error":   true,
			"data":    gin.H{},
			"message": fmt.Sprintf("Error: %v", res.Error),
			"status":  400,
		})

		// fmt.Printf("rrr %+v\n ", new_group)

		return
	}
	// send response
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"error":   false,
		"data": gin.H{
			"group":         new_group,
			"admin_details": admin,
		},
		"message": "Group created successfully",
		"status":  200,
	})
	// fmt.Printf("rrr %+v\n ", new_group)

}

// get user by email api
func GetMemberByEmail(c *gin.Context) {

	user_email := c.Param("user_email")
	// fmt.Println(user_email)

	// find user by email :
	var user models.User
	res := config.DB.Where("email=?", user_email).First(&user)
	fmt.Println(res)
	if res.Error != nil {
		if res.Error == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"data":    gin.H{},
				"error":   true,
				"message": "User with this email not found",
				"status":  404,
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": res.Error.Error()})
		}
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    user,
		"error":   true,
		"message": "User with this email found",
		"status":  200,
	})

}

// add  members by email api:
func AddGroupMemberByEmail(c *gin.Context) {

	var member_body struct {
		Member_email string `json:"member_email"`
		Group_id     int    `json:"groupId"`
	}

	if c.Bind(&member_body) != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Failed to read body",
		})
		return
	}

	// Get user ID from context
	user_id, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Find the user by email
	var user models.User
	if err := config.DB.Where("email = ?", member_body.Member_email).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   true,
			"data":    gin.H{},
			"message": "Error: User with this email not found",
			"status":  400,
		})
		return
	}

	// create a check if user already exists :
	var existingMember models.GroupMembers
	if err := config.DB.Where("group_id = ?  AND member_id = ?", member_body.Group_id, user.ID).First(&existingMember).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{
			"success": false,
			"error":   true,
			"data":    gin.H{},
			"message": "Error: Member already exists in this group",
			"status":  409,
		})
		return
	}

	new_member := models.GroupMembers{
		GroupId:      member_body.Group_id,
		Member_email: member_body.Member_email,
		MemberId:     int(user.ID),
	}

	res := config.DB.Create(&new_member)

	if res.Error != nil {
		fmt.Println(res.Error)
		c.JSON(http.StatusBadRequest, gin.H{

			"success": false,
			"error":   true,
			"data":    gin.H{},
			"message": fmt.Sprintf("Error: %v", res.Error),
			"status":  400,
		})

		return
	}

	// Fetch group details including group admin
	var group models.Group
	if err := config.DB.Preload("Admin").Where("id = ?", member_body.Group_id).First(&group).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   true,
			"data":    gin.H{},
			"message": "Error: Unable to fetch group details",
			"status":  500,
		})
		return
	}

	// send response
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"error":   false,
		"data": gin.H{
			"details": gin.H{
				"ID":          new_member.ID,
				"GroupId":     new_member.GroupId,
				"MemberId":    new_member.MemberId,
				"MemberEmail": new_member.Member_email,

				"Group": gin.H{
					"ID":             group.ID,
					"CreatedAt":      group.CreatedAt,
					"UpdatedAt":      group.UpdatedAt,
					"DeletedAt":      group.DeletedAt,
					"group_name":     group.Group_name,
					"description":    group.Description,
					"category":       group.Category,
					"group_admin_id": group.Group_admin_id,
				},
				"Admin": gin.H{
					"ID":       group.Admin.ID,
					"UserName": group.Admin.Name,
					"Email":    group.Admin.Email,
				},

				"user": gin.H{
					"userId": user_id,
				},
			},
		},
		"message": "Member added successfully",
		"status":  200,
	})
	// fmt.Printf("rrr %+v\n ", new_group)

}

func GetAllGroupsById(c *gin.Context) {

	user_id := c.Param("user_id")
	var group []models.Group
	// res := config.DB.Where("id=?", user_id).First(&user)
	err := config.DB.
		Preload("Admin").
		Joins("LEFT JOIN group_members ON groups.id = group_members.group_id").
		Where("groups.group_admin_id = ? OR group_members.member_id = ?", user_id, user_id).
		Group("groups.id, group_members.group_id").
		Find(&group).Error

	fmt.Println(err)
	// fmt.Println(group)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   true,
			"message": "No groups found",
			"status":  400,
			"data":    err,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"error":   false,
		"message": "All groups",
		"status":  200,
		"data": gin.H{
			"alldata": group,
			"total":   len(group),
		},
	})

}

func GetAllGroupMembersByGroupId(c *gin.Context) {
	group_id := c.Param("group_id")

	type GroupMemberResponse struct {
		MemberId    int    `json:"member_id"`
		MemberEmail string `json:"member_email"`
		IsAdmin     bool   `json:"is_admin"`
		MemberName  string `json:"member_name"`
	}

	var group_members []models.GroupMembers
	err := config.DB.Preload("User").Preload("Group").Preload("Group.Admin").Where("group_id=?", group_id).Find(&group_members).Error
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   true,
			"message": "No group members present",
			"status":  400,
			"data":    err,
		})
		return
	}
	// Check if no group members are found
	if len(group_members) == 0 {
		c.JSON(http.StatusOK, gin.H{
			"data":    group_members,
			"error":   false,
			"message": "No group members present",
			"status":  204,
			"success": true,
		})
		return
	}

	// getting group admin  :
	var group models.Group
	er := config.DB.Preload("Admin").Where("id = ?", group_id).First(&group).Error
	if er != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   true,
			"message": "Error fetching group admin",
			"status":  400,
			"data":    er,
		})
		return
	}

	// making response now :
	var response []GroupMemberResponse
	for _, ele := range group_members {
		member := GroupMemberResponse{
			MemberId:    int(ele.MemberId),
			MemberEmail: ele.Member_email,
			IsAdmin:     int(ele.MemberId) == int(group.Admin.ID),
			MemberName:  ele.User.Name,
		}
		response = append(response, member)
	}

	// adding group member to the list :
	adminMember := GroupMemberResponse{
		MemberId:    int(group.Admin.ID),
		MemberEmail: group.Admin.Email,
		IsAdmin:     true,
		MemberName:  group.Admin.Name,
	}
	response = append(response, adminMember)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"error":   false,
		"message": "Members fetched successfully",
		"status":  200,
		"data": gin.H{
			// "details": group_members,
			// "group_admin": group_members,
			"total":   len(group_members) + 1,
			"details": response,
		},
	})

}

func DeleteMemberFromGroup(c *gin.Context) {
	groupId := c.Param("group_id")
	memberId := c.Param("member_id")
	groupAdminId := c.Param("group_admin_id")
	loggedin_user_id := c.Param("loggedin_user_id")

	fmt.Println(memberId)
	fmt.Println(groupAdminId)
	fmt.Println(groupAdminId == memberId)
	var group_member models.GroupMembers
	var group models.Group

	// loggedin_user_id, exists := c.Get("user_id")

	// if !exists {
	// 	c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
	// 	return
	// }

	// fetcing group admin id from groups :
	if err := config.DB.Select("group_admin_id").Where("id = ?", groupId).First(&group).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"data":    gin.H{},
			"error":   true,
			"message": "error : Group not found",
			"status":  400,
			"success": false,
		})
		return
	}

	// groupAdminID := group.Group_admin_id

	// checking if member exists in group or not :
	if err := config.DB.Where("group_id = ? AND member_id = ?", groupId, memberId).First(&group_member).Error; err != nil {

		fmt.Println("new checkkkk heree ")

		fmt.Println(groupId)
		fmt.Println(memberId)
		fmt.Println(err)

		c.JSON(http.StatusNotFound, gin.H{
			"data":    gin.H{},
			"error":   true,
			"message": "error : Member not found in the group",
			"status":  400,
			"success": false,
		})
		return
	}

	// checking if the logged in user is the group admin or not :
	// new_lui := uint(loggedin_user_id)
	// userIDStr, ok := loggedin_user_id.(string)
	// if !ok {
	// 	c.JSON(400, gin.H{"error": "User ID is not of type string"})
	// 	return
	// }
	// Convert string to uint
	// uintUserID, err := strconv.ParseUint(userIDStr, 10, 64)
	// if err != nil {
	// 	c.JSON(400, gin.H{"error": "Failed to convert User ID to uint"})
	// 	return
	// }

	// new_gai := uint(groupAdminID)
	// fmt.Println(loggedin_user_id)
	// fmt.Println(groupAdminID)
	// fmt.Println(reflect.TypeOf(loggedin_user_id))
	// fmt.Println(reflect.TypeOf(groupAdminID))

	// fmt.Println(uintUserID == groupAdminID)
	// fmt.Println(new_gai == uintUserID)

	// Convert loggedinUserID to int
	// var loggedinUserIDInt int
	// switch reflect.TypeOf(loggedin_user_id).Kind() {

	// case reflect.Int:
	// 	loggedinUserIDInt = loggedin_user_id.(int)
	// case reflect.Uint:
	// 	loggedinUserIDInt = loggedin_user_id.(int)
	// case reflect.String:
	// 	var err error
	// 	loggedinUserIDInt, err = strconv.Atoi(loggedin_user_id.(string))
	// 	if err != nil {
	// 		c.JSON(400, gin.H{"error": "Failed to convert loggedinUserID to int"})
	// 		return
	// 	}
	// default:
	// 	c.JSON(400, gin.H{"error": "User ID is not of expected type"})
	// 	return
	// }

	if loggedin_user_id != groupAdminId {
		c.JSON(http.StatusNotFound, gin.H{
			"data":    gin.H{},
			"error":   true,
			"message": "Only group admin can delete group member",
			"status":  400,
			"success": false,
		})
		return
	}

	// deleting the member by group admin only:

	err := config.DB.Where("group_id = ? AND member_id = ?", groupId, memberId).Delete(&group_member).Error
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"data":    gin.H{},
			"error":   true,
			"message": "Failed to delete member",
			"status":  400,
			"success": false,
		})
		return

	}

	c.JSON(http.StatusOK, gin.H{
		"data":    gin.H{},
		"error":   false,
		"message": "Member deleted successfully",
		"status":  200,
		"success": true,
	})

}
