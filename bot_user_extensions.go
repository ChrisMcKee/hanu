package hanu

// User Helpers
func contains(arr []string, str string) bool {
	for _, a := range arr {
		if a == str {
			return true
		}
	}
	return false
}

func (b *Bot) GetUserName(conv Convo) (string, error) {

	api := b.SocketClient.Client
	user, err := api.GetUserInfo(conv.Message().User())
	if err != nil {
		return "", err
	}
	return user.Profile.RealName, nil
}

func (b *Bot) GetUserEmail(conv Convo) (string, error) {
	api := b.SocketClient.Client
	user, err := api.GetUserInfo(conv.Message().User())
	if err != nil {
		return "", err
	}
	return user.Profile.Email, nil
}

func (b *Bot) IsUserInGroup(email string, userGroup string, conv Convo) bool {
	api := b.SocketClient.Client

	userGroups, err := api.GetUserGroupMembers(userGroup)
	if err != nil {
		api.Debugf("Unexpected error: %s", err)
		return false
	}

	return contains(userGroups, email)
}
