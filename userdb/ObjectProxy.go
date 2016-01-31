package userdb

type (
	// A proxy that implements the Object interface. Can be useful for
	// exposing an Account or Object to CanAccess() without actually
	// looking up the id and thereby saving a database hit.
	ObjectProxy string
)

func (p ObjectProxy) GetAccountId() string {
	return string(p)
}
