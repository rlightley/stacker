package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"gopkg.in/yaml.v3"
	"io/ioutil"
)

type Config struct {
	Subscriptions []Subscription `yaml:"subscriptions"`
	Environments  []string       `yaml:"environments"`
	Regions       []string       `yaml:"regions"`
}

type Subscription struct {
	Name      string     `yaml:"name"`
	Resources []Resource `yaml:"resources"`
}

type Resource struct {
	Name        string       `yaml:"name"`
	ExcludeFrom ExcludeConfig `yaml:"exclude-from"`
}

type ExcludeConfig struct {
	Environments []string `yaml:"environments"`
	Regions      []string `yaml:"regions"`
}

func main() {
	config, err := loadConfig("config.yml")
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		return
	}

	for _, sub := range config.Subscriptions {
		subFolder := sub.Name
		if err := os.Mkdir(subFolder, 0755); err != nil && !os.IsExist(err) {
			fmt.Printf("Error creating subscription folder '%s': %v\n", subFolder, err)
			continue
		}

		for _, env := range config.Environments {
			envFolder := fmt.Sprintf("%s/%s", subFolder, env)
			if err := os.Mkdir(envFolder, 0755); err != nil && !os.IsExist(err) {
				fmt.Printf("Error creating environment folder '%s': %v\n", envFolder, err)
				continue
			}

			for _, region := range config.Regions {
				regionFolder := fmt.Sprintf("%s/%s", envFolder, region)
				if err := os.Mkdir(regionFolder, 0755); err != nil && !os.IsExist(err) {
					fmt.Printf("Error creating region folder '%s': %v\n", regionFolder, err)
					continue
				}

				for _, resource := range sub.Resources {
					if shouldSkip(resource, env, region) {
						continue
					}

					resourceFolder := fmt.Sprintf("%s/%s", regionFolder, resource.Name)
					if err := os.Mkdir(resourceFolder, 0755); err != nil && !os.IsExist(err) {
						fmt.Printf("Error creating resource folder '%s': %v\n", resourceFolder, err)
						continue
					}

					tags := strings.Join([]string{sub.Name, resource.Name, region, env}, ",")
					if err := runTerramateCommand(resourceFolder, tags); err != nil {
						fmt.Printf("Error running terramate command in '%s': %v\n", resourceFolder, err)
					}
				}
			}
		}
	}
}

func loadConfig(filename string) (*Config, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	return &config, nil
}

func shouldSkip(resource Resource, env, region string) bool {
	for _, excludedEnv := range resource.ExcludeFrom.Environments {
		if strings.EqualFold(env, excludedEnv) {
			return true
		}
	}
	for _, excludedRegion := range resource.ExcludeFrom.Regions {
		if strings.EqualFold(region, excludedRegion) {
			return true
		}
	}
	return false
}

func runTerramateCommand(folder, tags string) error {
	cmd := exec.Command("terramate", "create", "--tags", tags, folder)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	fmt.Printf("Running terramate command in folder: %s with tags: %s\n", folder, tags)
	return cmd.Run()
}
