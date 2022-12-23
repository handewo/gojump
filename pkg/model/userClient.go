package model

type UserClient struct {
	Username   string
	Password   string
	PublicKey  string
	RemoteAddr string
}

type UserClientOption func(*UserClient)

func UserClientPassword(password string) UserClientOption {
	return func(args *UserClient) {
		args.Password = password
	}
}

func UserClientPublicKey(publicKey string) UserClientOption {
	return func(args *UserClient) {
		args.PublicKey = publicKey
	}
}

func UserClientRemoteAddr(addr string) UserClientOption {
	return func(args *UserClient) {
		args.RemoteAddr = addr
	}
}

func (u *UserClient) SetOption(setters ...UserClientOption) {
	for _, setter := range setters {
		setter(u)
	}
}
