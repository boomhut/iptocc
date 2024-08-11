package iptocc

import (
	"fmt"
	"io/ioutil"
	"net"
	"strings"

	"github.com/ip2location/ip2location-go/v9"
)

type Ip2LocationDataFiles struct {
	DataFolder string // data folder path to store IP2Location data files. Example: /path/to/data/
	IPv4       string // IPv4 data file name. Example: IP2LOCATION-LITE-DB11.BIN
	IPv6       string // IPv6 data file name. Example: IP2LOCATION-LITE-DB11.IPV6.BIN
}

var ip2loc *Ip2LocationDataFiles

// Function to set IP2Location data folder
func SetDataFolder(dataFolder string) {
	ip2loc = new(Ip2LocationDataFiles)
	ip2loc.DataFolder = dataFolder

	i4, i6 := FindDataFiles()
	ip2loc.IPv4 = i4
	ip2loc.IPv6 = i6
}

// function to automatically find the IP2Location data files in the data folder
func FindDataFiles() (string, string) {
	// first find the IPv6 data file (ending with .IPV6.BIN)
	// loop through all files in the data folder
	files, err := ioutil.ReadDir(ip2loc.DataFolder)
	if err != nil {
		fmt.Println(err)
		return "", ""
	}
	for _, file := range files {
		if strings.HasSuffix(file.Name(), ".IPV6.BIN") {
			ip2loc.IPv6 = file.Name()
			break
		}
	}

	// find the IPv4 data file (ending with .BIN, but not .IPV6.BIN)
	// loop through all files in the data folder
	for _, file := range files {
		if strings.HasSuffix(file.Name(), ".BIN") && !strings.HasSuffix(file.Name(), ".IPV6.BIN") {
			ip2loc.IPv4 = file.Name()
			break
		}
	}

	return ip2loc.IPv4, ip2loc.IPv6
}

// // Function to set IP2Location data files
func init() {
	SetDataFolder("./data/")
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
	db, err := ip2location.OpenDB(ip2loc.DataFolder + ip2loc.IPv4)

	if err != nil {
		fmt.Print(err)
		return ipInfo{}
	}
	results, err := db.Get_all(ip)

	if err != nil {
		fmt.Print(err)
		return ipInfo{}
	}
	// lookup hostname
	hostname, err := net.LookupAddr(ip)
	if err != nil {
		fmt.Print(err)
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
	db, err := ip2location.OpenDB(ip2loc.DataFolder + ip2loc.IPv6)

	if err != nil {
		fmt.Print(err)
		return ipInfo{}
	}
	results, err := db.Get_all(ip)

	if err != nil {
		fmt.Print(err)
		return ipInfo{}
	}

	// lookup hostname
	hostname, err := net.LookupAddr(ip)
	if err != nil {
		fmt.Print(err)
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
