package notifier

type User struct {
	Username       string `json:"username"`
	Password       string `json:"password"`
	Token          string `json:"token"`
	AllowAnonymous bool   `json:"allowAnonymous"`
}
