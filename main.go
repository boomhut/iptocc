package iptocc

import (
	"fmt"
	"net"
	"strings"

	"github.com/ip2location/ip2location-go/v9"
)

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
	db, err := ip2location.OpenDB("./IP2LOCATION-LITE-DB11.BIN")

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
	db, err := ip2location.OpenDB("./IP2LOCATION-LITE-DB11.IPV6.BIN")

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
