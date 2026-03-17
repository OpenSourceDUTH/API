package team

type TeamMember struct {
	ID                 int      `json:"id"`
	Name               string   `json:"name"`
	Surname            string   `json:"surname"`
	Email              string   `json:"email"`
	GitHubUsername     string   `json:"github_username"`
	EmailNotifications bool     `json:"email_notifications"`
	TeamRoles          []string `json:"team_roles"`
	IsAdmin            bool     `json:"is_admin"`
	IsActive           bool     `json:"is_active"`
	Alumni             bool     `json:"alumni"`
}
