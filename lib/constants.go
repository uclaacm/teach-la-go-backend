package lib

// DEFAULT_USER_DATA
// Unfortunately, Go does not permit struct constants.
// We fix this problem by just writing a function that
// manufactures the default user struct for us.
func DEFAULT_USER_DATA() *UserData {
	u := UserData{DisplayName: "J Bruin"}
	return &u
}