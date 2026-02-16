package handlers

import "testing"

func TestParseYear(t *testing.T) {
    cases := []struct{
        in string
        want int
    }{
        {"1994", 1994},
        {"01/02/1980", 1980},
        {"first album 2001 edition", 2001},
        {"no year", 0},
        {"abcd2019ef", 2019},
    }
    for _, c := range cases {
        got := parseYear(c.in)
        if got != c.want {
            t.Fatalf("parseYear(%q) = %d; want %d", c.in, got, c.want)
        }
    }
}

func TestArtistMatchesLocations(t *testing.T) {
    artistLocs := []string{"Seattle, Washington, USA", "Paris, France"}
    cases := []struct{
        sel []string
        want bool
    }{
        {[]string{"Seattle, Washington, USA"}, true},
        {[]string{"Washington, USA"}, true},
        {[]string{"France"}, true},
        {[]string{"Berlin, Germany"}, false},
    }
    for _, c := range cases {
        got := artistMatchesLocations(artistLocs, c.sel)
        if got != c.want {
            t.Fatalf("artistMatchesLocations(%v, %v) = %v; want %v", artistLocs, c.sel, got, c.want)
        }
    }
}
