package iptocc

import (
	"encoding/gob"
	"fmt"
	"net"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/ip2location/ip2location-go/v9"
)

type Ip2LocationDataFiles struct {
	DataFolder string    // data folder path to store IP2Location data files. Example: /path/to/data/
	SubFolder  string    // sub folder path to store IP2Location data files. Example: /path/to/data/september2023/
	LastUpdate time.Time // last update time of sub folder.
	IPv4       string    // IPv4 data file name. Example: IP2LOCATION-LITE-DB11.BIN
	IPv6       string    // IPv6 data file name. Example: IP2LOCATION-LITE-DB11.IPV6.BIN
	mu         sync.RWMutex
}

var ip2loc *Ip2LocationDataFiles
var ip4DB *ip2location.DB
var ip6DB *ip2location.DB

// init function to create a new instance of Ip2LocationDataFiles, set the data folder and find the data files and connect to the database
func init() {
	ip2loc = new(Ip2LocationDataFiles)

	// db
	ip4DB = nil // set default IPv4 database to nil
	ip6DB = nil // set default IPv6 database to nil

	SetDataFolder("./data/ip2location")
	loadConfig()

}

// GetIp4DB function to get the IPv4 database
func GetIp4DB() *ip2location.DB {
	if ip4DB == nil {
		ipDB4, err := ip2location.OpenDB(ip2loc.DataFolder + ip2loc.IPv4)
		if err != nil {
			fmt.Println(err)
			return nil
		}
		ip4DB = ipDB4
	}
	return ip4DB
}

// GetIp6DB function to get the IPv6 database
func GetIp6DB() *ip2location.DB {
	if ip6DB == nil {
		ipDB6, err := ip2location.OpenDB(ip2loc.DataFolder + ip2loc.IPv6)
		if err != nil {
			fmt.Println(err)
			return nil
		}
		ip6DB = ipDB6
	}
	return ip6DB
}

// Function to set IP2Location data folder
func SetDataFolder(dataFolder string) error {
	// check if data folder exists
	if _, err := os.Stat(dataFolder); os.IsNotExist(err) {
		return fmt.Errorf("data folder does not exist: %s", dataFolder)
	}

	// if ip2loc is nil, create a new instance
	if ip2loc == nil {
		ip2loc = new(Ip2LocationDataFiles)
	}
	// check if data folder is empty
	if dataFolder == "" {
		return fmt.Errorf("data folder is empty")
	}
	// check if data folder is a directory
	if fi, err := os.Stat(dataFolder); err != nil || !fi.IsDir() {
		return fmt.Errorf("data folder is not a directory: %s", dataFolder)
	}

	// check if data folder is writable
	if err := os.MkdirAll(dataFolder, 0755); err != nil {
		return fmt.Errorf("data folder is not writable: %s", dataFolder)
	}

	ip2loc.mu.Lock()
	defer ip2loc.mu.Unlock()

	ip2loc.DataFolder = dataFolder

	return nil

}

func loadConfig() error {
	// lock struct for write
	ip2loc.mu.Lock()
	defer ip2loc.mu.Unlock()
	// check if data folder is empty
	if ip2loc.DataFolder == "" {
		return fmt.Errorf("data folder is empty")
	}
	// check if data folder is a directory
	if fi, err := os.Stat(ip2loc.DataFolder); err != nil || !fi.IsDir() {
		return fmt.Errorf("data folder is not a directory: %s", ip2loc.DataFolder)
	}

	// check if data folder is writable
	if err := os.MkdirAll(ip2loc.DataFolder, 0755); err != nil {
		return fmt.Errorf("data folder is not writable: %s", ip2loc.DataFolder)
	}

	// check if config file exists (ip2l_config.bin)
	if _, err := os.Stat(ip2loc.DataFolder + "/ip2l_config.bin"); os.IsNotExist(err) {
		// create config file
		f, err := os.Create(ip2loc.DataFolder + "/ip2l_config.bin")
		if err != nil {
			return fmt.Errorf("error creating config file: %s", err)
		}
		defer f.Close()
		// write config file. config file is a gob file containing the ip2location data files
		// Ip2LocationDataFiles is a struct containing the data folder, sub folder, IPv4 and IPv6 data files
		// Gob encode the Ip2LocationDataFiles struct and write it to the config file
		enc := gob.NewEncoder(f)
		err = enc.Encode(ip2loc)
		if err != nil {
			return fmt.Errorf("error encoding config file: %s", err)
		}
	} else {
		// read config file
		f, err := os.Open(ip2loc.DataFolder + "/ip2l_config.bin")
		if err != nil {
			return fmt.Errorf("error opening config file: %s", err)
		}
		defer f.Close()
		// decode config file. config file is a gob file containing the ip2location data files
		// Ip2LocationDataFiles is a struct containing the data folder, sub folder, IPv4 and IPv6 data files
		// Gob decode the Ip2LocationDataFiles struct and read it from the config file
		dec := gob.NewDecoder(f)
		err = dec.Decode(ip2loc)
		if err != nil {
			return fmt.Errorf("error decoding config file: %s", err)
		}

	}

	return nil
}

// OnExit function to close the database connections and save the config file
func OnExit() {
	// close the database connections
	if ip4DB != nil {
		ip4DB.Close()
		ip4DB = nil
	}
	if ip6DB != nil {
		ip6DB.Close()
		ip6DB = nil
	}

	// save the config file
	if err := saveConfig(); err != nil {
		fmt.Println(err)
		return
	}
}

// saveConfig function to save the config file (ip2l_config.bin) to the data folder
func saveConfig() error {
	// check if data folder is empty
	if ip2loc.DataFolder == "" {
		return fmt.Errorf("data folder is empty")
	}

	// check if data folder is a directory
	if fi, err := os.Stat(ip2loc.DataFolder); err != nil || !fi.IsDir() {
		return fmt.Errorf("data folder is not a directory: %s", ip2loc.DataFolder)
	}

	// check if data folder is writable
	if err := os.MkdirAll(ip2loc.DataFolder, 0755); err != nil {
		return fmt.Errorf("data folder is not writable: %s", ip2loc.DataFolder)
	}

	// save config file. config file is a gob file containing the ip2location data files
	// Ip2LocationDataFiles is a struct containing the data folder, sub folder, IPv4 and IPv6 data files
	// Gob encode the Ip2LocationDataFiles struct and write it to the config file
	f, err := os.Create(ip2loc.DataFolder + "ip2l_config.bin")
	if err != nil {
		return fmt.Errorf("error creating config file: %s", err)
	}

	defer f.Close()
	enc := gob.NewEncoder(f)
	err = enc.Encode(ip2loc)
	if err != nil {
		return fmt.Errorf("error encoding config file: %s", err)
	}
	return nil

}

// Function to set IP2Location data folder
func SetSubFolder(dataFolder string) error {
	// check if data folder exists
	if _, err := os.Stat(dataFolder); os.IsNotExist(err) {
		return fmt.Errorf("data folder does not exist: %s", dataFolder)
	}

	// if ip2loc is nil, create a new instance
	if ip2loc == nil {
		ip2loc = new(Ip2LocationDataFiles)
	}
	// check if data folder is empty
	if dataFolder == "" {
		return fmt.Errorf("data folder is empty")
	}
	// check if data folder is a directory
	if fi, err := os.Stat(dataFolder); err != nil || !fi.IsDir() {
		return fmt.Errorf("data folder is not a directory: %s", dataFolder)
	}

	// check if data folder is writable
	if err := os.MkdirAll(dataFolder, 0755); err != nil {
		return fmt.Errorf("data folder is not writable: %s", dataFolder)
	}

	ip2loc.mu.Lock()
	defer ip2loc.mu.Unlock()

	i4, i6, err := FindDataFiles(dataFolder)
	if err != nil {
		return fmt.Errorf("error finding data files: %s", err)
	}

	// set the data folder path to store IP2Location data files
	ip2loc.SubFolder = dataFolder

	// set the sub folder path to store IP2Location data files
	ip2loc.IPv4 = i4
	ip2loc.IPv6 = i6

	if ip2loc.IPv4 == "" || ip2loc.IPv6 == "" {
		return fmt.Errorf("data files not found in data folder: %s", dataFolder)
	}

	return nil
}

// updateSubFolder function to update the sub folder path to store IP2Location data files
func updateSubFolder(dataFolder string) error {
	// check if data folder exists
	if _, err := os.Stat(dataFolder); os.IsNotExist(err) {
		return fmt.Errorf("data folder does not exist: %s", dataFolder)
	}

	// if ip2loc is nil, create a new instance
	if ip2loc == nil {
		ip2loc = new(Ip2LocationDataFiles)
	}
	// check if data folder is empty
	if dataFolder == "" {
		return fmt.Errorf("data folder is empty")
	}
	// check if data folder is a directory
	if fi, err := os.Stat(dataFolder); err != nil || !fi.IsDir() {
		return fmt.Errorf("data folder is not a directory: %s", dataFolder)
	}

	ip2loc.mu.Lock()
	// set the sub folder path to store IP2Location data files
	ip2loc.SubFolder = dataFolder
	ip2loc.LastUpdate = time.Now()
	ip2loc.mu.Unlock()

	return nil
}

// function to automatically find the IP2Location data files in the data folder
func FindDataFiles(subFolder string) (string, string, error) {
	var ipv4, ipv6 string
	// first find the IPv6 data file (ending with .IPV6.BIN)
	// loop through all files in the data folder
	files, err := os.ReadDir(subFolder)
	if err != nil {
		return "", "", fmt.Errorf("error reading data folder: %s", err)
	}
	for _, file := range files {
		if strings.HasSuffix(file.Name(), ".IPV6.BIN") {
			ipv6 = file.Name()
			break
		}
	}

	// find the IPv4 data file (ending with .BIN, but not .IPV6.BIN)
	// loop through all files in the data folder
	for _, file := range files {
		if strings.HasSuffix(file.Name(), ".BIN") && !strings.HasSuffix(file.Name(), ".IPV6.BIN") {
			ipv4 = file.Name()
			break
		}
	}

	// check if both files are found
	if ipv4 == "" || ipv6 == "" {
		return "", "", fmt.Errorf("data files not found in data folder: %s", ip2loc.SubFolder)
	}

	return ipv4, ipv6, nil
}

// Function to lookup country by IP address
func LookupCountry(ip net.IP) (string, error) {

	// check if IP is valid
	if ip == nil {
		return "", fmt.Errorf("invalid IP address")
	}

	// check if IP is a loopback address
	if ip.IsLoopback() {
		return "Loopback address", nil
	}

	// check if IP is a private address
	if ip.IsPrivate() {
		return "Private address", nil
	}

	// ip 4 or 6
	switch {
	case ip.To4() != nil:
		return Ip4ToLocation(ip.String()).Country_short, nil
	case ip.To16() != nil:
		return Ip6ToLocation(ip.String()).Country_short, nil

	default:
		return "", fmt.Errorf("invalid IP address")
	}

}

type ipInfo struct {
	Address       string  // ip address
	Hostname      string  // hostname
	Type          string  // ipv4 or ipv6
	Country_short string  // country short name
	Country_long  string  // country long name
	Region        string  // region
	City          string  // city
	Latitude      float32 // latitude
	Longitude     float32 // longitude
	Zipcode       string  // zipcode
	Timezone      string  // timezone
	Elevation     float32 // elevation
}

// ipInfo to string
func (info ipInfo) String() string {
	return fmt.Sprintf("Address: %s\nHostname: %s\nCountry_short: %s\nCountry_long: %s\nRegion: %s\nCity: %s\nLatitude: %f\nLongitude: %f\nZipcode: %s\nTimezone: %s\nElevation: %f\n", info.Address, info.Hostname, info.Country_short, info.Country_long, info.Region, info.City, info.Latitude, info.Longitude, info.Zipcode, info.Timezone, info.Elevation)
}

func Ip4ToLocation(ip string) ipInfo {
	// lock ip2location database for read
	// ip2loc.mu.RLock()
	// defer ip2loc.mu.RUnlock()
	// db, err := ip2location.OpenDB(ip2loc.DataFolder + ip2loc.IPv4)

	db := GetIp4DB()
	if db == nil {
		fmt.Println("Error opening IPv4 database")
		return ipInfo{}
	}

	results, err := db.Get_all(ip)

	if err != nil {
		// fmt.Print(err)
		return ipInfo{}
	}
	// lookup hostname
	hostname, err := net.LookupAddr(ip)
	if err != nil {
		// fmt.Print(err)
		return ipInfo{}
	}

	return ipInfo{
		Address:       ip,
		Hostname:      strings.Join(hostname, "."),
		Country_short: results.Country_short,
		Country_long:  results.Country_long,
		Region:        results.Region,
		City:          results.City,
		Latitude:      results.Latitude,
		Longitude:     results.Longitude,
		Zipcode:       results.Zipcode,
		Timezone:      results.Timezone,
		Elevation:     results.Elevation,
	}

}

func Ip6ToLocation(ip string) ipInfo {
	// lock ip2location database for read
	// ip2loc.mu.RLock()
	// defer ip2loc.mu.RUnlock()
	// db, err := ip2location.OpenDB(ip2loc.DataFolder + ip2loc.IPv6)

	// if err != nil {
	// 	// fmt.Print(err)
	// 	return ipInfo{}
	// }
	db := GetIp6DB()
	if db == nil {
		fmt.Println("Error opening IPv6 database")
		return ipInfo{}
	}
	results, err := db.Get_all(ip)

	if err != nil {
		// fmt.Print(err)
		return ipInfo{}
	}

	// lookup hostname
	hostname, err := net.LookupAddr(ip)
	if err != nil {
		// fmt.Print(err)
		return ipInfo{}
	}

	return ipInfo{
		Address:       ip,
		Hostname:      strings.Join(hostname, "."),
		Country_short: results.Country_short,
		Country_long:  results.Country_long,
		Region:        results.Region,
		City:          results.City,
		Latitude:      results.Latitude,
		Longitude:     results.Longitude,
		Zipcode:       results.Zipcode,
		Timezone:      results.Timezone,
		Elevation:     results.Elevation,
	}
}

// Ip2Location
func Ip2Location(ip string) ipInfo {
	// check if IP is valid ipv4 or ipv6
	if net.ParseIP(ip) == nil {
		return ipInfo{}
	}

	// ip 4 or 6
	switch {
	case net.ParseIP(ip).To4() != nil:
		return Ip4ToLocation(ip)
	case net.ParseIP(ip).To16() != nil:
		return Ip6ToLocation(ip)
	default:
		return ipInfo{}
	}
}
