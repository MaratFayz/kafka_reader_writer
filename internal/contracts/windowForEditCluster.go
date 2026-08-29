package contracts

type CloseWindowForEditCluster struct{}

type ErrorSaveEditedClusterIntoLocalStorageEvent struct {
	Err error
}

type SuccessSaveEditedClusterIntoLocalStorageEvent struct {
}

type UserConfirmEditClusterInList struct {
	Title            string
	Password         string
	Trust_store_path string
	Url              string
	Username         string
}
