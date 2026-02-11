package config

func GetDefault() *Config {
	return &Config{
		SearchLimit:         DefaultSearchLimit,
		DefaultDownloadPath: DefaultDownloadPath,
		DefaultFormat:       DefaultFormat,
		SortByDefault:       DefaultSortBy,
		EmbedSubtitles:      DefaultEmbedSubtitles,
		EmbedMetadata:       DefaultEmbedMetadata,
		EmbedChapters:       DefaultEmbedChapters,
		CookiesBrowser:      DefaultCookiesBrowser,
		CookiesFile:         DefaultCookiesFile,
	}
}

const DefaultSearchLimit = 25

const DefaultDownloadPath = "~/Downloads"

const DefaultFormat = "bestvideo+bestaudio/best"

const DefaultSortBy = "relevance"

const DefaultEmbedSubtitles = false

const DefaultEmbedMetadata = true

const DefaultEmbedChapters = true

const DefaultCookiesBrowser = ""

const DefaultCookiesFile = ""
