package contracts

type UserRefuseAddClusterToList struct{}

type UserConfirmAddNewClusterToList struct {
	Title            string
	Url              string
	Username         string
	Password         string
	Trust_store_path string
}

type SuccessSaveNewClusterIntoLocalStorageEvent struct {
}

type ErrorSaveNewClusterIntoLocalStorageEvent struct {
	Err error
}
