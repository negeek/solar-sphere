package main

import (
    "fmt"
    "os"
    "os/exec"
    "path/filepath"
	"plugin"
)


var MigrationsDir string =  "./db_migrations/migration_files"
var MigratedDir string = "./db_migrations/migrated_files"

func Migrate() error {
    // This function is to go to the migrations folder and run the MakeMigration() functions in each file.
    return filepath.Walk(MigrationsDir, func(path string, info os.FileInfo, err error) error {
        if err != nil {
            fmt.Println(err.Error())
            return err
        }
        if !info.IsDir() && filepath.Ext(path) == ".go" {
            migrated, err := SkipAlreadyMigrated(path)
            if err != nil {
                fmt.Println(err.Error())
                return err
            }
            if !migrated {
                fmt.Printf("Migrating : %s\n", path)
                if err := executeMakeMigrationFunction(path); err != nil {
                    fmt.Println(err.Error())
                    return err
                }
                fmt.Printf("Migrated: %s\n", path)
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
    makeMigrationFunc, err := p.Lookup("Add")
    if err != nil {
        return err
    }

    // Call the MakeMigration function
    if makeMigrationFunc != nil {
        if fn, ok := makeMigrationFunc.(func()); ok {
            fn()
        } else {
            return fmt.Errorf("MakeMigration is not a function")
        }
    } else {
        return fmt.Errorf("MakeMigration function not found in %s", filePath)
    }

    return nil
}

// compileGoFile compiles the given Go file and returns the path to the compiled binary
func compileGoFile(filePath string) (string, error) {
    // Define the output directory for compiled binaries
    if err := os.MkdirAll(MigratedDir, 0755); err != nil {
        return "", err
    }
	outDir:=filepath.Dir(MigratedDir)
    // Define the output file name for the compiled binary
    outFile := filepath.Join(outDir, filepath.Base(filePath))
    // Run the 'go build' command to compile the Go file into a binary
    cmd := exec.Command("go", "build", "-buildmode=plugin", "-o", outFile, filePath)
    cmd.Stderr = os.Stderr
    if err := cmd.Run(); err != nil {
        return "", fmt.Errorf("failed to compile %s: %v", filePath, err)
    }

    return outFile, nil
}

func SkipAlreadyMigrated(filePath string) (bool, error) {
	migratedDir := filepath.Dir(MigratedDir)
    compiledFilePath := filepath.Join(migratedDir, fileName)
    // Check if the file exists
    if _, err := os.Stat(compiledFilePath); err == nil {
        fmt.Printf("File %s exists in directory %s\n", fileName, dir)
        return true, nil
    } else if os.IsNotExist(err) {
        fmt.Printf("File %s does not exist in directory %s\n", fileName, dir)
        return false, nil
    } else {
        fmt.Printf("Error checking file existence: %v\n", err)
        return false, err
    }
}

func main(){
	err:= Migrate()
	if err != nil{
		fmt.Println(err)
	}
}