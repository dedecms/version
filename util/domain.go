package util

import (
	"strings"

	"github.com/dedecms/snake"
	"golang.org/x/net/idna"
)

// HasSubdomain reports whether domain contains any subdomain.
func HasSubdomain(domain string) bool {
	domain, top := stripURLParts(domain), Domain(domain)
	return domain != top && top != ""
}

// Subdomain returns subdomain from provided url.
// If subdomain is not found in provided url, this function returns empty string.
func Subdomain(url string) string {
	domain, top := stripURLParts(url), Domain(url)
	lt, ld := len(top), len(domain)
	if lt < ld && top != "" {
		return domain[:(ld-lt)-1]
	}
	return ""
}

// SplitDomain split domain into string array
// for example, zh.wikipedia.org will split into {"zh", "wikipedia", "org"}
func SplitDomain(url string) []string {
	domain, second, top := Subdomain(url), DomainPrefix(url), DomainSuffix(url)
	if len(top) == 0 {
		return nil
	}

	if len(second) == 0 {
		return []string{top}
	}

	if len(domain) == 0 {
		return []string{second, top}
	}

	array := strings.Split(domain, ".")
	res := append(array, second, top)
	return res
}

// DomainPrefix returns second-level domain from provided url.
// If no SLD is found in provided url, this function returns empty string.
func DomainPrefix(url string) string {
	domain := Domain(url)
	if len(domain) != 0 {
		return domain[:strings.Index(domain, ".")]
	}
	return ""
}

// Host returns Host from provided url.
func Host(url string) string {
	return stripURLParts(url)
}

// DomainSuffix returns domain suffix from provided url.
// If no TLD is found in provided url, this function returns empty string.
func DomainSuffix(url string) string {
	domain := Domain(url)
	if len(domain) != 0 {
		return domain[strings.Index(domain, ".")+1:]
	}
	return ""
}

// Domain returns top level domain from url string.
// If no domain is found in provided url, this function returns empty string.
// If no TLD is found in provided url, this function returns empty string.
func Domain(url string) string {

	domain, top := stripURLParts(url), ""
	parts := strings.Split(domain, ".")
	currentTld := *tlds
	foundTld := false

	// Cycle trough parts in reverse
	if len(parts) > 1 {
		for i := len(parts) - 1; i >= 0; i-- {
			// Generate top domain output
			if top != "" {
				top = "." + top
			}
			top = parts[i] + top

			// Check for TLD
			if currentTld == nil {
				return top // Return current output because we no longer have the TLD
			} else if tldEntry, found := currentTld[parts[i]]; found {
				if tldEntry != nil {
					currentTld = *tldEntry
				} else {
					currentTld = nil
				}
				foundTld = true
				continue
			} else if foundTld {
				return top // Return current output if tld was found before
			}

			// Return empty string if no tld was found ever
			return ""
		}
	}

	return ""
}

func URL(url string) string {
	// Strip protocol
	if index := strings.Index(url, "://"); index > -1 {
		url = url[index+3:]
	}
	return snake.String(url).Trim("/").Get()
}

// stripURLParts removes path, protocol & query from url and returns it.
func stripURLParts(url string) string {

	// Lower case the url
	url = strings.ToLower(url)

	// Strip protocol
	if index := strings.Index(url, "://"); index > -1 {
		url = url[index+3:]
	}

	// Now, if the url looks like this: username:password@www.example.com/path?query=?
	// we remove the content before the '@' symbol
	if index := strings.Index(url, "@"); index > -1 {
		url = url[index+1:]
	}

	// Strip path (and query with it)
	if index := strings.Index(url, "/"); index > -1 {
		url = url[:index]
	} else if index := strings.Index(url, "?"); index > -1 { // Strip query if path is not found
		url = url[:index]
	}

	if colon := strings.LastIndexByte(url, ':'); colon != -1 && validOptionalPort(url[colon:]) {
		url = url[:colon]
	}

	if strings.HasPrefix(url, "[") && strings.HasSuffix(url, "]") {
		url = url[1 : len(url)-1]
	}

	// Convert domain to unicode
	if strings.Index(url, "xn--") != -1 {
		var err error
		url, err = idna.ToUnicode(url)
		if err != nil {
			return ""
		}
	}

	// Return domain
	return url
}

// validOptionalPort reports whether port is either an empty string
// or matches /^:\d*$/
func validOptionalPort(port string) bool {
	if port == "" {
		return true
	}
	if port[0] != ':' {
		return false
	}
	for _, b := range port[1:] {
		if b < '0' || b > '9' {
			return false
		}
	}
	return true
}

// Protocol returns protocol from given url
//
// If protocol is not present - return empty string
func Protocol(url string) string {
	if index := strings.Index(url, "://"); index > -1 {
		return url[:index]
	}
	return ""
}

// credentials returns credentials (user:pass) from given url
func credentials(url string) string {
	index := strings.IndexRune(url, '@')
	if index == -1 {
		return ""
	}
	if protocol := Protocol(url); protocol != "" {
		return url[len(protocol)+3 : index]
	}
	return url[:index]
}

// Username returns username from given url
//
// If username is not present - return empty string
func Username(url string) string {
	auth := strings.SplitN(credentials(url), ":", 2)
	if len(auth) == 0 {
		return ""
	}
	return auth[0]
}

// Password returns password from given url
//
// If password is not present - return empty string
func Password(url string) string {
	auth := strings.SplitN(credentials(url), ":", 2)
	if len(auth) < 2 {
		return ""
	}
	return auth[1]
}
