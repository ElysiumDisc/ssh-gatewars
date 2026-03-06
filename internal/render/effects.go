package render

// Explosion frame patterns for reference:
// Frame 0: *           (bright red)
// Frame 1: \|/ -X- /|\ (orange)
// Frame 2: .:. :.:     (warm orange fade)
// Frame 3: . .         (dim gray)
//
// Duration: ~400ms total (4 frames x ~100ms each at 10 ticks/sec)

// ExplosionColors maps explosion frame to color.
var ExplosionColors = []string{"#FF4444", "#FF8844", "#FFAA44", "#666666"}

// TrailChars are the dim trail markers behind moving ships.
const TrailChar = "·"
