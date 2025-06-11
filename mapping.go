package main

import "strings"

// Event name mapping from Mixpanel to PostHog
var nameMapping = map[string]string{
	// Apply page events
	"typeformSubmissionError":    "applicant_apply--typeform_submission_error",
	"completedApplicantTypeform": "applicant_apply--typeform_submission_success",

	// Auth events
	"applicantSignUp": "applicant_sign_up",

	// Message events
	"firstMessageSent": "first_message_send",

	// Session events
	"coachingSessionScheduled": "coaching_session_scheduled",
	"introCallScheduled":       "intro_call_scheduled",

	// Subscription events
	"subscriptionStarted":   "manage_subscription--subscription_start",
	"subscriptionCanceled":  "manage_subscription--subscription_cancel",

	// Purchase events
	"purchase":              "purchase_complete",
	"completeGuestCheckout": "checkout--guest_checkout_complete",
	"completeGuestSignup":   "checkout--guest_signup_complete",

	// Event card events
	"Click Event Card Register - Navigate to Luma": "event_card--navigate_to_luma_click",
	"EventCard Click":                              "event_card--click",

	// Vouch events
	"Vouch Modal Open":    "vouch_modal--open",
	"Vouch Modal Submit":  "vouch_modal--submit",
	"Vouch Outcome Submit": "vouch_outcome_step--submit",

	// Events page events
	"EventsPage - click - Class": "events_page--class_card_click",

	// Coach events
	"viewedIntroVideo":        "coach_profile--intro_video_view",
	"viewedCoachProfile":      "coach_profile--view",
	"clickedMessageCoach":     "coach_message_cta--message_coach_click",
	"coachMessageCTA - messaged coach": "coach_message_cta--message_coach_success",

	// Schedule modal events
	"SCHEDULE_MODAL_EVENTS.STEP_CHANGE": "schedule_modal--step_change",
	"SCHEDULE_MODAL_EVENTS.CLOSE":       "schedule_modal--close",

	// Search events
	"clickedSRPPackage":                    "srp--leland_package_click",
	"SRP - click - Search Bar":             "srp--search_bar_focus",
	"SRP - filter - Search Bar":            "srp--search_bar_filter",
	"SRP - click - Featured Filter":        "srp--featured_filter_click",
	"SRP - click - Booked Coach":           "srp--booked_coach_click",
	"Cohort Banner Click":                  "srp--cohort_banner_click",
	"SRP - click - Class":                  "srp--class_click",
	"SRP - click - Leland Package":         "srp--leland_package_click",
	"SRP - click - View more packages":     "srp--view_more_packages_click",
	"SRP - click - Offering Package":       "srp--offering_package_click",
	"SRP - click - Sort":                   "srp--coach_filter_section_sort",
	"SRP - click - Coach Card":             "srp--coach_card_click",
	"SRP - click - Coach Pagination":       "srp--coach_pagination_click",
	"SRP - click - Talk to a Team Member button": "srp--talk_to_a_team_member_button_click",

	// Class events
	"Enroll Free Event":   "free_event--enroll",
	"Unenroll Free Event": "free_event--unenroll",

	// Meeting events
	"coachingSessionAttended": "meeting--coaching_session_attend",

	// Bootcamp events
	"Bootcamp Card Click": "bootcamp_card--click",

	// Article events
	"Article - Subscribe":                "article_page--email_list_subscribe",
	"Article - Schedule a strategy call": "article_page--schedule_call_button_click",

	// Event Banner
	"Event Banner Click": "article_page--event_banner_click",

	// Page View
	"pageView": "$pageview",
}

// MapEventName maps a Mixpanel event name to its PostHog equivalent
func MapEventName(event string, properties map[string]interface{}) string {
	// Handle special case for Leland+ Banner Click
	if event == "Leland+ Banner Click" {
		if source, ok := properties["source"].(string); ok {
			switch strings.ToLower(source) {
			case "srp":
				return "srp--leland_plus_banner_click"
			case "article":
				return "article_page--leland_plus_banner_click"
			case "post_checkout":
				return "post_checkout--leland_plus_banner_click"
			default:
				return "leland_plus_banner--click"
			}
		}
		return "leland_plus_banner--click"
	}

	// Handle regular mappings
	if mapped, ok := nameMapping[event]; ok {
		return mapped
	}
	return event
} 