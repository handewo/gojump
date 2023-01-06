package model

type TerminalConfig struct {
	AssetListPageSize string `json:"TERMINAL_ASSET_LIST_PAGE_SIZE"`
	HeaderTitle       string `json:"TERMINAL_HEADER_TITLE"`
	HostKey           string `json:"-"`
	//Minute
	MaxIdleTime        int  `json:"MAX_IDLE_TIME"`
	AssetListSortByIp  bool `json:"TERMINAL_ASSET_LIST_SORT_BY_IP"`
	PasswordAuth       bool `json:"TERMINAL_PASSWORD_AUTH"`
	PublicKeyAuth      bool `json:"TERMINAL_PUBLIC_KEY_AUTH"`
	EnableSessionShare bool `json:"SECURITY_SESSION_SHARE"`
}
