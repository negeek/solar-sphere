package main

import (
    "errors"
    "log"
    "os"
    "os/exec"
    "path/filepath"
	"plugin"
	"runtime"
	"strings"
)

var MigrationDir string =  "/db_migrations/migration_files"
var MigratedDir string = "/db_migrations/migrated_files"
var TrackChanges int = 0

func getThisFileDir() string {
    // Get the caller's PC (program counter)
    pc, _, _, ok := runtime.Caller(1)
    if !ok {
		log.Fatal("Unable to get caller information")
        return ""
    }

    // Get the function name and file path of the caller
    funcInfo := runtime.FuncForPC(pc)
    if funcInfo == nil {
		log.Fatal("Unable to get caller function information")
        return ""
    }

    // Retrieve the file name of the caller
    fullPath, _ := funcInfo.FileLine(pc)

    // Exclude this file itself in the path
	fullPathArr := strings.Split(fullPath,"/")
	Path := strings.Join(fullPathArr[0:len(fullPathArr)-1],"/")

    return Path
}

func Migrate() error {
	migrationDir:=filepath.Join(getThisFileDir(), MigrationDir)
    // This function is to go to the migrations folder and run the MakeMigration() functions in each file.
    return filepath.Walk(migrationDir, func(path string, info os.FileInfo, err error) error {
        if err != nil {
            log.Fatal(err)
            return err
        }
        if !info.IsDir() && filepath.Ext(path) == ".go" {
            migrated, err := SkipAlreadyMigrated(path)
            if err != nil {
                log.Fatal(err)
                return err
            }
            if !migrated {
                TrackChanges+=1
                log.Printf("Migrating: %s\n", filepath.Base(path))
                if err := executeMakeMigrationFunction(path); err != nil {
                    log.Fatal(err)
                    return err
                }
                log.Printf("Migrated: %s\n", filepath.Base(path))
            }
        }
        return nil
    })
}

func executeMakeMigrationFunction(filePath string) error {
    // Load the compiled Go file as a plugin
    pluginPath, err := compileGoFile(filePath)
    if err != nil {
        return err
    }
    p, err := plugin.Open(pluginPath)
    if err != nil {
        return err
    }

    // Look up the MakeMigration symbol
    makeMigrationFunc, err := p.Lookup("MakeMigration")
    if err != nil {
        return err
    }

    // Call the MakeMigration function
    if makeMigrationFunc != nil {
        if fn, ok := makeMigrationFunc.(func()); ok {
            fn()
        } else {
            return errors.New("MakeMigration is not a function")
        }
    } else {
        return errors.New("MakeMigration function not found")
    }

    return nil
}

func compileGoFile(filePath string) (string, error) {
    // Define the output directory for compiled binaries
	migratedDir:= filepath.Join(getThisFileDir(), MigratedDir)
    if err := os.MkdirAll(migratedDir, 0755); err != nil {
        return "", err
    }
    // Define the output file name for the compiled binary
    outFile := filepath.Join(migratedDir, filepath.Base(filePath))
    // Run the 'go build' command to compile the Go file into a binary
    cmd := exec.Command("go", "build", "-buildmode=plugin", "-o", outFile, filePath)
    cmd.Stderr = os.Stderr
    if err := cmd.Run(); err != nil {
        return "", errors.New("Failed to compile")
    }

    return outFile, nil
}

func SkipAlreadyMigrated(filePath string) (bool, error) {
	migratedDir:= filepath.Join(getThisFileDir(), MigratedDir)
	fileName:=filepath.Base(filePath)
    compiledFilePath := filepath.Join(migratedDir, fileName)
    // Check if the file exists
    if _, err := os.Stat(compiledFilePath); err == nil {
        return true, nil
    } else if os.IsNotExist(err) {
        return false, nil
    } else {
        return false, err
    }
}

func main(){
	err := Migrate()
	if err != nil{
		log.Fatal(err)
	}
    if TrackChanges < 1 {
        log.Println("No changes made")
    }
}