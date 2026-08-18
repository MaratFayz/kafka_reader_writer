package debug

import "os"

func WriteDebugFile(text string) {
	file, err := os.OpenFile("file.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	if _, err := file.WriteString(text + "\n"); err != nil {
		panic(err)
	}
}
