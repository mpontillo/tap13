package tap13

import (
	"bufio"
	"log"
	"os"
)

func ReadFile(name string) []string {
	file, err := os.Open(name)

	if err != nil {
		log.Fatalf("Could not open file: %s", err)
	}

	defer file.Close()

	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)
	var lines []string

	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	return lines
}
