package controller

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

// CheckAPIHandler 返回一组json用来给当前账号授予一些网页版的视觉上面的特性
func CheckAPIHandler(c *gin.Context) {
	accountInfo := gin.H{
		"account": gin.H{
			"account_user_role": "account-owner",
			"account_user_id":   "user-000000000000000000000000__a323bd05-db25-4e8f-9173-2f0c228cc8fa",
			"processor": gin.H{
				"a001": gin.H{
					"has_customer_object": true,
				},
				"b001": gin.H{
					"has_transaction_history": false,
				},
				"c001": gin.H{
					"has_transaction_history": false,
				},
			},
			"account_id":      "a323bd05-db25-4e8f-9173-2f0c228cc8fa",
			"organization_id": nil,
			"is_most_recent_expired_subscription_gratis": false,
			"has_previously_paid_subscription":           true,
			"name":                                       nil,
			"structure":                                  "personal",
			"promo_data":                                 gin.H{},
		},
		"features": []string{
			"priority_driven_models_list",
			"browsing_publisher_red_team",
			"plugin_review_tools",
			"message_debug_info",
			"user_latency_tools",
			"tools3_dev",
			"debug",
			"workspace_share_links",
			"retrieval_poll_ui",
			"sunshine_available",
			"use_stream_processor",
			"voice_available",
			"i18n",
			"model_switcher",
			"persist_last_used_model",
			"code_interpreter_available",
			"breeze_available",
			"beta_features",
			"starter_prompts",
			"browsing_available",
			"new_plugin_oauth_endpoint",
			"dalle_3",
			"layout_may_2023",
			"shareable_links",
			"allow_url_thread_creation",
			"invite_referral",
			"plugins_available",
			"ks",
			"chat_preferences_available",
			"model_switcher_upsell",
		},
		"entitlement": gin.H{
			"subscription_id":         "d0dcb1fc-56aa-4cd9-90ef-37f1e03576d3",
			"has_active_subscription": true,
			"subscription_plan":       "chatgptplusplan",
			"expires_at":              "2089-08-08T23:59:59+00:00",
		},
		"last_active_subscription": gin.H{
			"subscription_id":          "d0dcb1fc-56aa-4cd9-90ef-37f1e03576d3",
			"purchase_origin_platform": "chatgpt_web",
			"will_renew":               true,
		},
	}
	// 下面这组数据是从Pandora中直接拿出来的
	data := gin.H{
		"accounts": gin.H{
			"a323bd05-db25-4e8f-9173-2f0c228cc8fa": accountInfo,
			"default":                              accountInfo,
		},
		"account_ordering": []string{
			"a323bd05-db25-4e8f-9173-2f0c228cc8fa",
		},
	}
	c.JSON(http.StatusOK, data)
}
