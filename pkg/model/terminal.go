package model

type TerminalConfig struct {
	AssetListPageSize  string `json:"TERMINAL_ASSET_LIST_PAGE_SIZE"`
	AssetListSortByIp  bool   `json:"TERMINAL_ASSET_LIST_SORT_BY_IP"`
	HeaderTitle        string `json:"TERMINAL_HEADER_TITLE"`
	PasswordAuth       bool   `json:"TERMINAL_PASSWORD_AUTH"`
	PublicKeyAuth      bool   `json:"TERMINAL_PUBLIC_KEY_AUTH"`
	MaxIdleTime        int    `json:"MAX_IDLE_TIME"`
	HostKey            string `json:"TERMINAL_HOST_KEY"`
	EnableSessionShare bool   `json:"SECURITY_SESSION_SHARE"`
}
