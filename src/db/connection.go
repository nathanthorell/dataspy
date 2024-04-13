package db

type Connection struct {
	Type     string `json:"Type"`
	Server   string `json:"Server"`
	Port     string `json:"Port"`
	Database string `json:"Database"`
	User     string `json:"User"`
	Password string `json:"Password"`
}
