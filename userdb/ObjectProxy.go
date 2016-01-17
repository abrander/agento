package userdb

type (
	ObjectProxy string
)

func (p ObjectProxy) GetAccountId() string {
	return string(p)
}
