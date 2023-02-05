package main

import "log"
import "strings"
import "errors"
import "fmt"
import "unicode"
import "regexp"

type State struct {
	sayRate int
	speechRate int
	uppercasePitch float32
}

func newState() *State {
	s := State{}
	s.sayRate = 550
	s.speechRate = 275
	s.uppercasePitch = 0.8
	return &s
}

var state = newState()
var queue = []string{}
var embedCommand = regexp.MustCompile(`\[\{.*?\}\]`)
	
func processLine(l string) error {
	if debugMode() {
		log.Println("Processing: ", l)
	}
	command, body := getParts(l)
	switch command {
	case "version":
		handleVersion()
	case "tts_say":
		handleTtsSay(body)
	case "l":
		handleLetter(body)
	case "d":
		handleDispatch()
	case "c":
		handleCode(body)
	case "q":
		handleText(body)
	case "s":
		handleStop()
	default:
		return errors.New("Unknown command: "+ command)
	}
	
	return nil
}

func handleText(s string) {
	queue = append(queue, s)
}

func handleCode(s string) {
	queue = append(queue, s)
}

func handleDispatch() {
	if len(queue) > 0 {
		emitVoice(strings.Join(queue, ""))
	}
}

func handleVersion() {
	emitVoice(withRate(fmt.Sprintf("Version is $s", version), state.speechRate))
}

func handleStop() {
	queue = []string{}
	speaking, err := NsSpeechIsSpeaking()
	if err != nil {
		if debugMode() {
			log.Println("Error checking if speaking: ", err)
		}
	}
	if speaking {
		err = NsSpeechStop()
		if err != nil {
			if debugMode() {
				log.Println("Error stopping: ", err)
			}
		}
	}

}

func handleTtsSay(s string) {
	emitVoice(withRate(s, state.speechRate))
}

func handleLetter(s string) {
	var first rune
	for _,c := range s {
		first = c
		break
	}

	if unicode.IsUpper(first) {
		s = wrapPitch(s, state.uppercasePitch)
	}

	s = wrapChar(s)
	s = withRate(s, state.sayRate)
		
	emitVoice(s)
}

func withRate(s string, rate int) string {
	return fmt.Sprintf(`[[rate %d]]%s`, rate, s)
}

func wrapChar(s string) string {
	return fmt.Sprintf(`[[char ltr]]%s[[char norm]]`, s)
}

func wrapPitch(s string, pitch float32) string {
	return fmt.Sprintf(`[[pbas +%f]]%s[[pbas -%f]]`, pitch, s, pitch)
}
	
func getParts(s string) (string, string) {
	r := strings.SplitN(strings.Trim(s, " "), " ", 2)
	if len(r) == 1 {
		return r[0], ""
	}
	return r[0], replaceMorpheme(removeBraceWrappers(r[1]))
}

func replaceMorpheme(s string) string {
	return strings.Replace(s, `[*]`, `[[slnc 50]]`, -1)
}

func removeBraceWrappers(s string) string {
	i := strings.Index(s, "{")
	li := strings.LastIndex(s, "}")
	if i != -1 && li != -1 {
		return s[i+1:li-1]
	}
	return s
}

func emitVoice(t string) {
	t = embedCommand.ReplaceAllString(t, "")
	err := NsSpeechSpeak(t)
	if err != nil {
		if debugMode() {
			log.Println("failed to emitVoice: ", err)
		}
	}
}
