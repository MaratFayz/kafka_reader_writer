package debug

import "os"

func WriteDebugFile(text string) {
	// Открываем файл в режиме добавления
	// O_APPEND - добавлять в конец
	// O_CREATE - создать если не существует
	// O_WRONLY - только для записи
	file, err := os.OpenFile("file.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	// Записываем строку
	if _, err := file.WriteString(text + "\n"); err != nil {
		panic(err)
	}
}
