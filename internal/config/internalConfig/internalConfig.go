// internalConfig contains not so volatile parameters used in the internals 
package internalConfig


// >>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>
// Configurations for the internal notification system.
type NotifsConfig struct {
	GetNotifsLimit int32 // number of notifs to return per request (page size)
}
func LoadNotifsConfig() NotifsConfig {
	return NotifsConfig{
		GetNotifsLimit: 15,
	}
}

// >>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>

// Configurations for discussions page.
type DiscussionConfig struct {
	NewPostUpperLimit int // maximum number of characters allowed for a new post 
	NewPostLowerLimit int // minimum number of characters required for a new post
}
func LoadDiscussionConfig() DiscussionConfig {
	return DiscussionConfig{
		NewPostUpperLimit: 2500,
		NewPostLowerLimit: 50,
	}
}

// >>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>

// Configurations for feedbacks page.
type FeedbacksConfig struct {
	NewMessageUpperLimit int // maximum number of characters allowed for a new post 
	NewMessageLowerLimit int // minimum number of characters required for a new post
}
func LoadFeedbacksConfig() FeedbacksConfig {
	return FeedbacksConfig{
		NewMessageUpperLimit: 500,
		NewMessageLowerLimit: 25,
	}
}