package main

import "time"

// CountryCodeToTimezone maps phone country codes to their primary timezone
// For countries with multiple timezones, we use the most common/populous one
var CountryCodeToTimezone = map[string]string{
	// Asia
	"+91":  "Asia/Kolkata",      // India
	"+92":  "Asia/Karachi",      // Pakistan
	"+93":  "Asia/Kabul",        // Afghanistan
	"+94":  "Asia/Colombo",      // Sri Lanka
	"+95":  "Asia/Yangon",       // Myanmar
	"+960": "Indian/Maldives",   // Maldives
	"+971": "Asia/Dubai",        // UAE
	"+972": "Asia/Jerusalem",    // Israel
	"+973": "Asia/Bahrain",      // Bahrain
	"+974": "Asia/Qatar",        // Qatar
	"+975": "Asia/Thimphu",      // Bhutan
	"+976": "Asia/Ulaanbaatar",  // Mongolia
	"+977": "Asia/Kathmandu",    // Nepal
	"+98":  "Asia/Tehran",       // Iran
	"+880": "Asia/Dhaka",        // Bangladesh
	"+81":  "Asia/Tokyo",        // Japan
	"+82":  "Asia/Seoul",        // South Korea
	"+84":  "Asia/Ho_Chi_Minh",  // Vietnam
	"+86":  "Asia/Shanghai",     // China
	"+852": "Asia/Hong_Kong",    // Hong Kong
	"+853": "Asia/Macau",        // Macau
	"+855": "Asia/Phnom_Penh",   // Cambodia
	"+856": "Asia/Vientiane",    // Laos
	"+60":  "Asia/Kuala_Lumpur", // Malaysia
	"+62":  "Asia/Jakarta",      // Indonesia (most populous timezone)
	"+63":  "Asia/Manila",       // Philippines
	"+65":  "Asia/Singapore",    // Singapore
	"+66":  "Asia/Bangkok",      // Thailand

	// Europe
	"+44":  "Europe/London",      // UK
	"+33":  "Europe/Paris",       // France
	"+49":  "Europe/Berlin",      // Germany
	"+34":  "Europe/Madrid",      // Spain
	"+39":  "Europe/Rome",        // Italy
	"+31":  "Europe/Amsterdam",   // Netherlands
	"+32":  "Europe/Brussels",    // Belgium
	"+41":  "Europe/Zurich",      // Switzerland
	"+43":  "Europe/Vienna",      // Austria
	"+45":  "Europe/Copenhagen",  // Denmark
	"+46":  "Europe/Stockholm",   // Sweden
	"+47":  "Europe/Oslo",        // Norway
	"+48":  "Europe/Warsaw",      // Poland
	"+351": "Europe/Lisbon",      // Portugal
	"+353": "Europe/Dublin",      // Ireland
	"+354": "Atlantic/Reykjavik", // Iceland
	"+358": "Europe/Helsinki",    // Finland
	"+359": "Europe/Sofia",       // Bulgaria
	"+36":  "Europe/Budapest",    // Hungary
	"+370": "Europe/Vilnius",     // Lithuania
	"+371": "Europe/Riga",        // Latvia
	"+372": "Europe/Tallinn",     // Estonia
	"+380": "Europe/Kyiv",        // Ukraine
	"+7":   "Europe/Moscow",      // Russia (Moscow timezone as default)
	"+30":  "Europe/Athens",      // Greece
	"+40":  "Europe/Bucharest",   // Romania
	"+420": "Europe/Prague",      // Czech Republic
	"+421": "Europe/Bratislava",  // Slovakia
	"+386": "Europe/Ljubljana",   // Slovenia
	"+385": "Europe/Zagreb",      // Croatia
	"+381": "Europe/Belgrade",    // Serbia
	"+90":  "Europe/Istanbul",    // Turkey

	// Americas
	"+1":   "America/New_York",     // USA/Canada (Eastern as default - most populous)
	"+52":  "America/Mexico_City",  // Mexico
	"+55":  "America/Sao_Paulo",    // Brazil
	"+54":  "America/Buenos_Aires", // Argentina
	"+56":  "America/Santiago",     // Chile
	"+57":  "America/Bogota",       // Colombia
	"+58":  "America/Caracas",      // Venezuela
	"+51":  "America/Lima",         // Peru
	"+593": "America/Guayaquil",    // Ecuador
	"+591": "America/La_Paz",       // Bolivia
	"+595": "America/Asuncion",     // Paraguay
	"+598": "America/Montevideo",   // Uruguay
	"+506": "America/Costa_Rica",   // Costa Rica
	"+507": "America/Panama",       // Panama

	// Caribbean & NANP codes (must be checked BEFORE +1 due to longer prefix)
	"+1242": "America/Nassau",        // Bahamas
	"+1246": "America/Barbados",      // Barbados
	"+1264": "America/Anguilla",      // Anguilla
	"+1268": "America/Antigua",       // Antigua and Barbuda
	"+1284": "America/Tortola",       // British Virgin Islands
	"+1340": "America/St_Thomas",     // US Virgin Islands
	"+1345": "America/Cayman",        // Cayman Islands
	"+1441": "Atlantic/Bermuda",      // Bermuda
	"+1473": "America/Grenada",       // Grenada
	"+1649": "America/Grand_Turk",    // Turks and Caicos
	"+1658": "America/Jamaica",       // Jamaica (alt)
	"+1664": "America/Montserrat",    // Montserrat
	"+1670": "Pacific/Saipan",        // Northern Mariana Islands
	"+1671": "Pacific/Guam",          // Guam
	"+1684": "Pacific/Pago_Pago",     // American Samoa
	"+1721": "America/Lower_Princes", // Sint Maarten
	"+1758": "America/St_Lucia",      // Saint Lucia
	"+1767": "America/Dominica",      // Dominica
	"+1784": "America/St_Vincent",    // Saint Vincent
	"+1787": "America/Puerto_Rico",   // Puerto Rico
	"+1809": "America/Santo_Domingo", // Dominican Republic
	"+1829": "America/Santo_Domingo", // Dominican Republic (alt)
	"+1849": "America/Santo_Domingo", // Dominican Republic (alt)
	"+1868": "America/Port_of_Spain", // Trinidad and Tobago
	"+1869": "America/St_Kitts",      // Saint Kitts and Nevis
	"+1876": "America/Jamaica",       // Jamaica
	"+1939": "America/Puerto_Rico",   // Puerto Rico (alt)

	// Oceania
	"+61":  "Australia/Sydney",     // Australia (Sydney as default - most populous)
	"+64":  "Pacific/Auckland",     // New Zealand
	"+679": "Pacific/Fiji",         // Fiji
	"+675": "Pacific/Port_Moresby", // Papua New Guinea

	// Africa
	"+20":  "Africa/Cairo",         // Egypt
	"+27":  "Africa/Johannesburg",  // South Africa
	"+234": "Africa/Lagos",         // Nigeria
	"+254": "Africa/Nairobi",       // Kenya
	"+212": "Africa/Casablanca",    // Morocco
	"+216": "Africa/Tunis",         // Tunisia
	"+213": "Africa/Algiers",       // Algeria
	"+255": "Africa/Dar_es_Salaam", // Tanzania
	"+256": "Africa/Kampala",       // Uganda
	"+233": "Africa/Accra",         // Ghana
	"+225": "Africa/Abidjan",       // Ivory Coast
	"+221": "Africa/Dakar",         // Senegal
	"+251": "Africa/Addis_Ababa",   // Ethiopia

	// Middle East
	"+966": "Asia/Riyadh",  // Saudi Arabia
	"+962": "Asia/Amman",   // Jordan
	"+961": "Asia/Beirut",  // Lebanon
	"+964": "Asia/Baghdad", // Iraq
	"+965": "Asia/Kuwait",  // Kuwait
	"+968": "Asia/Muscat",  // Oman
}

// GetTimezoneFromPhone returns the timezone based on the phone number's country code
// Returns "UTC" if the country code is not found
func GetTimezoneFromPhone(phoneNumber string) string {
	if phoneNumber == "" {
		return "UTC"
	}

	// Try matching from longest to shortest prefix (for codes like +1, +44, +880, etc.)
	// Check 4-digit codes first, then 3-digit, then 2-digit, then 1-digit
	for length := 4; length >= 1; length-- {
		if len(phoneNumber) > length {
			prefix := phoneNumber[:length+1] // +1 for the '+' sign
			if tz, ok := CountryCodeToTimezone[prefix]; ok {
				return tz
			}
		}
	}

	return "UTC"
}

// IsValidTimezone checks if a timezone string is valid
func IsValidTimezone(tz string) bool {
	if tz == "" {
		return false
	}
	_, err := time.LoadLocation(tz)
	return err == nil
}

// GetUserLocalTime returns the current time in the user's timezone
func GetUserLocalTime(timezone string) time.Time {
	loc, err := time.LoadLocation(timezone)
	if err != nil {
		loc = time.UTC
	}
	return time.Now().In(loc)
}
