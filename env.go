package permissionbus

var tokenSecretKey []byte

func SetTokenSecretKey(s string) {
	b := []byte(s)
	if len(b) <= 32 {
		panic("at least 32 byte")
	}
	tokenSecretKey = b
}
