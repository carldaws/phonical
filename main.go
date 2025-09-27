package main

import (
	"embed"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/faiface/beep"
	"github.com/faiface/beep/mp3"
	"github.com/faiface/beep/speaker"
	"github.com/faiface/beep/wav"
	hook "github.com/robotn/gohook"
)

//go:embed sounds/*
var soundFiles embed.FS

var phonicsMap = map[rune]string{
	'a': "a.wav",
	'b': "b.wav",
	'c': "c.wav",
	'd': "d.wav",
	'e': "e.wav",
	'f': "f.wav",
	'g': "g.wav",
	'h': "h.wav",
	'i': "i.wav",
	'j': "j.wav",
	'k': "k.wav",
	'l': "l.wav",
	'm': "m.wav",
	'n': "n.wav",
	'o': "o.wav",
	'p': "p.wav",
	'q': "q.wav",
	'r': "r.wav",
	's': "s.wav",
	't': "t.wav",
	'u': "u.wav",
	'v': "v.wav",
	'w': "w.wav",
	'x': "x.wav",
	'y': "y.wav",
	'z': "z.wav",
}

var (
	speakerInitialized bool
	playQueue          = make(chan string, 100)
	verbose            = false
	soundCache         = make(map[string]*beep.Buffer)
	soundCacheMutex    sync.RWMutex
)

func initSpeaker() error {
	if speakerInitialized {
		return nil
	}

	format := beep.Format{
		SampleRate:  44100,
		NumChannels: 2,
		Precision:   2,
	}

	// Use a smaller buffer size for lower latency
	err := speaker.Init(format.SampleRate, format.SampleRate.N(time.Second/60))
	if err != nil {
		return fmt.Errorf("failed to initialize speaker: %w", err)
	}

	speakerInitialized = true
	return nil
}

func loadSound(soundPath string) (*beep.Buffer, beep.Format, error) {
	soundCacheMutex.RLock()
	if buffer, exists := soundCache[soundPath]; exists {
		soundCacheMutex.RUnlock()
		return buffer, beep.Format{SampleRate: 44100, NumChannels: 2, Precision: 2}, nil
	}
	soundCacheMutex.RUnlock()

	file, err := soundFiles.Open("sounds/" + soundPath)
	if err != nil {
		return nil, beep.Format{}, err
	}
	defer file.Close()

	var streamer beep.StreamSeekCloser
	var format beep.Format

	if strings.HasSuffix(soundPath, ".mp3") {
		streamer, format, err = mp3.Decode(file)
	} else if strings.HasSuffix(soundPath, ".wav") {
		streamer, format, err = wav.Decode(file)
	} else {
		return nil, beep.Format{}, fmt.Errorf("unsupported format: %s", soundPath)
	}

	if err != nil {
		return nil, beep.Format{}, err
	}
	defer streamer.Close()

	buffer := beep.NewBuffer(format)
	buffer.Append(streamer)

	soundCacheMutex.Lock()
	soundCache[soundPath] = buffer
	soundCacheMutex.Unlock()

	return buffer, format, nil
}

func playSound(soundPath string) {
	buffer, _, err := loadSound(soundPath)
	if err != nil {
		if verbose {
			log.Printf("Failed to load sound %s: %v", soundPath, err)
		}
		return
	}

	if !speakerInitialized {
		if err := initSpeaker(); err != nil {
			if verbose {
				log.Printf("Failed to initialize speaker: %v", err)
			}
			return
		}
	}

	streamer := buffer.Streamer(0, buffer.Len())
	done := make(chan bool)
	speaker.Play(beep.Seq(streamer, beep.Callback(func() {
		done <- true
	})))
	<-done
}

func soundPlayer() {
	for soundFile := range playQueue {
		playSound(soundFile)
	}
}

func handleKeyPress(char rune) {
	if soundFile, exists := phonicsMap[char]; exists {
		if verbose {
			fmt.Printf("Key pressed: %c - Playing: %s\n", char, soundFile)
		}

		select {
		case playQueue <- soundFile:
		default:
			if verbose {
				log.Println("Sound queue full, skipping")
			}
		}
	}
}

func preloadSounds() {
	if verbose {
		fmt.Println("Preloading sounds...")
	}

	for _, soundFile := range phonicsMap {
		_, _, err := loadSound(soundFile)
		if err != nil && verbose {
			log.Printf("Failed to preload %s: %v", soundFile, err)
		}
	}

	if verbose {
		fmt.Printf("Preloaded %d sounds\n", len(soundCache))
	}
}

func main() {
	if len(os.Args) > 1 && (os.Args[1] == "-v" || os.Args[1] == "--verbose") {
		verbose = true
	}

	if len(os.Args) > 1 && (os.Args[1] == "-h" || os.Args[1] == "--help") {
		fmt.Println("Phonical - A phonics learning tool for kids")
		fmt.Println("\nUsage:")
		fmt.Printf("  %s [options]\n", filepath.Base(os.Args[0]))
		fmt.Println("\nOptions:")
		fmt.Println("  -v, --verbose    Show verbose output")
		fmt.Println("  -h, --help       Show this help message")
		fmt.Println("\nPress ESC or Ctrl+C to exit")
		os.Exit(0)
	}

	fmt.Println("Phonical - Phonics Learning Tool")
	fmt.Println("System-wide phonics - works across all applications!")
	fmt.Println("Press Ctrl+C to exit")
	fmt.Println("\nNote: You may need to grant Accessibility permissions in:")
	fmt.Println("System Preferences → Security & Privacy → Privacy → Accessibility")

	// Initialize speaker first
	if err := initSpeaker(); err != nil {
		log.Fatal("Failed to initialize audio:", err)
	}

	// Preload all sounds for faster playback
	preloadSounds()

	go soundPlayer()

	// Set up signal handling
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Start the event hook
	evChan := hook.Start()
	defer hook.End()

	fmt.Println("\nListening for keystrokes system-wide...")

	for {
		select {
		case ev := <-evChan:
			if verbose {
				fmt.Printf("Event: Kind=%d, Rawcode=%d, Keychar=%d, Keycode=%d\n", ev.Kind, ev.Rawcode, ev.Keychar, ev.Keycode)
			}
			// gohook uses Kind 3 for key down events
			if ev.Kind == 3 {
				// Use the Keychar field which gives us the actual character
				if ev.Keychar != 0 {
					char := rune(ev.Keychar)
					// Convert to lowercase for our map
					char = rune(strings.ToLower(string(char))[0])
					handleKeyPress(char)
				} else if verbose {
					fmt.Printf("Non-character key: rawcode=%d\n", ev.Rawcode)
				}
			}
		case <-sigChan:
			fmt.Println("\nExiting Phonical...")
			return
		}
	}
}