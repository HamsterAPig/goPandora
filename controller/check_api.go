package controller

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

// CheckAPIHandler 返回一组json用来给当前账号授予一些网页版的视觉上面的特性
func CheckAPIHandler(c *gin.Context) {
	// 下面这组数据是从Pandora中直接拿出来的
	data := gin.H{
		"accounts": gin.H{
			"default": gin.H{
				"account": gin.H{
					"account_user_role": "account-owner",
					"account_user_id":   "d0322341-7ace-4484-b3f7-89b03e82b927",
					"processor": gin.H{
						"a001": gin.H{
							"has_customer_object": true,
						},
						"b001": gin.H{
							"has_transaction_history": true,
						},
					},
					"account_id": "a323bd05-db25-4e8f-9173-2f0c228cc8fa",
					"is_most_recent_expired_subscription_gratis": true,
					"name":      nil,
					"structure": "personal",
				},
				"features": []string{
					"model_switcher",
					"system_message",
					"priority_driven_models_list",
					"message_style_202305",
					"layout_may_2023",
					"plugins_available",
					"beta_features",
					"infinite_scroll_history",
					"browsing_inner_monologue",
					"new_plugin_oauth_endpoint",
					"code_interpreter_available",
					"chat_preferences_available",
					"plugin_review_tools",
					"message_debug_info",
					"shareable_links",
					"tools3_dev",
					"tools2",
					"debug",
				},
				"entitlement": gin.H{
					"subscription_id":         "d0dcb1fc-56aa-4cd9-90ef-37f1e03576d3",
					"has_active_subscription": true,
					"subscription_plan":       "chatgptplusplan",
					"expires_at":              "2089-08-08T23:59:59+00:00",
				},
				"last_active_subscription": gin.H{
					"subscription_id":          "d0dcb1fc-56aa-4cd9-90ef-37f1e03576d3",
					"purchase_origin_platform": "chatgpt_mobile_ios",
					"will_renew":               true,
				},
			},
		},
		"temp_ap_available_at": "2023-05-20T17:30:00+00:00",
	}
	c.JSON(http.StatusOK, data)
}
