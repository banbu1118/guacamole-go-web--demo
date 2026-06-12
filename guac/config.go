package guac

type Config struct {
	ConnectionID        string
	Protocol            string
	Parameters          map[string]string
	OptimalScreenWidth  int
	OptimalScreenHeight int
	OptimalResolution   int
	AudioMimetypes      []string
	VideoMimetypes      []string
	ImageMimetypes      []string
}

func NewGuacamoleConfiguration() *Config {
	return &Config{
		Parameters:          map[string]string{},
		OptimalScreenWidth:  1024,
		OptimalScreenHeight: 768,
		OptimalResolution:   96,
		AudioMimetypes:      make([]string, 0, 1),
		VideoMimetypes:      make([]string, 0, 1),
		ImageMimetypes:      make([]string, 0, 1),
	}
}
