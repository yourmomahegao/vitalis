package scss

import (
	"errors"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/bep/godartsass/v2"
)

var compilerSource string = "./cmd/scss/bin/dart-sass/sass"
var compilerReleases string = "https://github.com/sass/dart-sass/releases"

func CompileSCSS(inputFolder string, outputFolder string) error {
	log.Println("Starting SCSS files compilation...")

	_, err := os.Stat(compilerSource)
	if err != nil {
		log.Printf("Error starting compilation: %v not exists", err)
		log.Printf("You can download compiler from: %s", compilerReleases)
		return err
	}

	transpilerOptions := godartsass.Options{DartSassEmbeddedFilename: compilerSource}
	transpiler, err := godartsass.Start(transpilerOptions)

	if err != nil {
		log.Println("Error while processing SCSS files.")
		return err
	}

	compilationErrorAmount := 0

	filepath.WalkDir(inputFolder, func(path string, d fs.DirEntry, err error) error {
		outputPath := strings.Replace(path, inputFolder, outputFolder, 1)
		outputPathLen := len(outputPath)
		outputPathSplitted := strings.Split(outputPath, ".")
		outputPathSplittedLen := len(outputPathSplitted)

		if outputPathLen > 1 && outputPathSplittedLen > 1 && outputPathSplitted[outputPathSplittedLen-1] == "scss" {
			outputPath = outputPath[:outputPathLen-4] + "css"
		}

		if err != nil {
			log.Printf("Error while directory walking: %v", err)
			compilationErrorAmount += 1
			return err
		}

		if d.IsDir() {
			log.Printf("Creting directory: %s", outputPath)
			err := os.MkdirAll(outputPath, 0755)

			if err != nil {
				if !errors.Is(err, os.ErrExist) {
					log.Printf("Failed during directory creation: %v", err)
					compilationErrorAmount += 1
				}
			}
		} else {
			if !strings.HasSuffix(path, ".scss") {
				return nil
			}

			log.Printf("Compiling file: %s -> %s", path, outputPath)

			scssData, err := os.ReadFile(path)
			if err != nil {
				log.Printf("Failed reading file: %v", err)
				compilationErrorAmount += 1
				return err
			}

			cssData, err := transpiler.Execute(godartsass.Args{Source: string(scssData)})
			if err != nil {
				log.Printf("Failed compiling file: %v", err)
				compilationErrorAmount += 1
				return err
			}

			err = os.WriteFile(outputPath, []byte(cssData.CSS), 0644)
			if err != nil {
				log.Printf("Failed writing file: %v", err)
				compilationErrorAmount += 1
				return err
			}
		}

		return err
	})

	log.Printf("SCSS compilation finished with %d error(s).", compilationErrorAmount)

	if compilationErrorAmount > 0 {
		return fmt.Errorf("Error while compiling SCSS files")
	}

	return nil
}
