package spamc

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"net"
	"regexp"
	"strconv"
	"strings"
	"time"
)

//Error Types

const EX_OK = 0           //no problems
const EX_USAGE = 64       // command line usage error
const EX_DATAERR = 65     // data format error
const EX_NOINPUT = 66     // cannot open input
const EX_NOUSER = 67      // addressee unknown
const EX_NOHOST = 68      // host name unknown
const EX_UNAVAILABLE = 69 // service unavailable
const EX_SOFTWARE = 70    // internal software error
const EX_OSERR = 71       // system error (e.g., can't fork)
const EX_OSFILE = 72      // critical OS file missing
const EX_CANTCREAT = 73   // can't create (user) output file
const EX_IOERR = 74       // input/output error
const EX_TEMPFAIL = 75    // temp failure; user is invited to retry
const EX_PROTOCOL = 76    // remote error in protocol
const EX_NOPERM = 77      // permission denied
const EX_CONFIG = 78      // configuration error
const EX_TIMEOUT = 79     // read timeout

//map of default spamD protocol errors v1.5
var SpamDError = map[int]string{
	EX_USAGE:       "Command line usage error",
	EX_DATAERR:     "Data format error",
	EX_NOINPUT:     "Cannot open input",
	EX_NOUSER:      "Addressee unknown",
	EX_NOHOST:      "Host name unknown",
	EX_UNAVAILABLE: "Service unavailable",
	EX_SOFTWARE:    "Internal software error",
	EX_OSERR:       "System error",
	EX_OSFILE:      "Critical OS file missing",
	EX_CANTCREAT:   "Can't create (user) output file",
	EX_IOERR:       "Input/output error",
	EX_TEMPFAIL:    "Temp failure; user is invited to retry",
	EX_PROTOCOL:    "Remote error in protocol",
	EX_NOPERM:      "Permission denied",
	EX_CONFIG:      "Configuration error",
	EX_TIMEOUT:     "Read timeout",
}

//Default parameters
const PROTOCOL_VERSION = "1.5"
const DEFAULT_TIMEOUT = 10

//Command types
const CHECK = "CHECK"
const SYMBOLS = "SYMBOLS"
const REPORT = "REPORT"
const REPORT_IGNOREWARNING = "REPORT_IGNOREWARNING"
const REPORT_IFSPAM = "REPORT_IFSPAM"
const SKIP = "SKIP"
const PING = "PING"
const TELL = "TELL"
const PROCESS = "PROCESS"
const HEADERS = "HEADERS"

//Learn types
const LEARN_SPAM = "SPAM"
const LEARN_HAM = "HAM"
const LEARN_NOTSPAM = "NOTSPAM"
const LEARN_NOT_SPAM = "NOT_SPAM"
const LEARN_FORGET = "FORGET"

//Test Types
const TEST_INFO = "info"
const TEST_BODY = "body"
const TEST_RAWBODY = "rawbody"
const TEST_HEADER = "header"
const TEST_FULL = "full"
const TEST_URI = "uri"
const TEST_TXT = "text"

//only for parse use !important
const SPLIT = "§"
const TABLE_MARK = "----"

//Types
type Client struct {
	ConnTimoutSecs  int
	ProtocolVersion string
	Host            string
	User            string
}

//Default response struct
type SpamDOut struct {
	Code    int
	Message string
	Vars    map[string]interface{}
}

//Default callback to SpanD response
type FnCallback func(*bufio.Reader) (*SpamDOut, error)

//new instance of Client
func New(host string, timeout int) *Client {
	return &Client{timeout, PROTOCOL_VERSION, host, ""}
}

func (s *Client) SetUnixUser(user string) {
	s.User = user
}

// Return a confirmation that spamd is alive.
func (s *Client) Ping() (r *SpamDOut, err error) {
	return s.simpleCall(PING, []string{})
}

// Just check if the passed message is spam or not and return score
func (s *Client) Check(msgpars ...string) (reply *SpamDOut, err error) {
	return s.simpleCall(CHECK, msgpars)
}

//Ignore this message -- client opened connection then changed its mind
func (s *Client) Skip(msgpars ...string) (reply *SpamDOut, err error) {
	return s.simpleCall(SKIP, msgpars)
}

//Check if message is spam or not, and return score plus list of symbols hit
func (s *Client) Symbols(msgpars ...string) (reply *SpamDOut, err error) {
	return s.simpleCall(SYMBOLS, msgpars)
}

//Check if message is spam or not, and return score plus report
func (s *Client) Report(msgpars ...string) (reply *SpamDOut, err error) {
	return s.simpleCall(REPORT, msgpars)
}

//Check if message is spam or not, and return score plus report
func (s *Client) ReportIgnoreWarning(msgpars ...string) (reply *SpamDOut, err error) {
	return s.simpleCall(REPORT_IGNOREWARNING, msgpars)
}

//Check if message is spam or not, and return score plus report if the message is spam
func (s *Client) ReportIfSpam(msgpars ...string) (reply *SpamDOut, err error) {
	return s.simpleCall(REPORT_IFSPAM, msgpars)
}

//Process this message and return a modified message - on deloy
func (s *Client) Process(msgpars ...string) (reply *SpamDOut, err error) {
	return s.simpleCall(PROCESS, msgpars)
}

//Same as PROCESS, but return only modified headers, not body (new in protocol 1.4)
func (s *Client) Headers(msgpars ...string) (reply *SpamDOut, err error) {
	return s.simpleCall(HEADERS, msgpars)
}

//Sign the message as spam
func (s *Client) ReportingSpam(msgpars ...string) (reply *SpamDOut, err error) {
	headers := map[string]string{
		"Message-class": "spam",
		"Set":           "local,remote",
	}
	return s.Tell(msgpars, &headers)
}

//Sign the message as false-positive
func (s *Client) RevokeSpam(msgpars ...string) (reply *SpamDOut, err error) {
	headers := map[string]string{
		"Message-class": "ham",
		"Set":           "local,remote",
	}
	return s.Tell(msgpars, &headers)
}

//Learn if a message is spam or not
func (s *Client) Learn(learnType string, msgpars ...string) (reply *SpamDOut, err error) {
	headers := make(map[string]string)
	switch strings.ToUpper(learnType) {
	case LEARN_SPAM:
		headers["Message-class"] = "spam"
		headers["Set"] = "local"
	case LEARN_HAM, LEARN_NOTSPAM, LEARN_NOT_SPAM:
		headers["Message-class"] = "ham"
		headers["Set"] = "local"
	case LEARN_FORGET:
		headers["Remove"] = "local"
	default:
		err = errors.New("Learn Type Not Found")
		return
	}
	return s.Tell(msgpars, &headers)
}

//wrapper to simple calls
func (s *Client) simpleCall(cmd string, msgpars []string) (reply *SpamDOut, err error) {
	return s.call(cmd, msgpars, func(data *bufio.Reader) (r *SpamDOut, e error) {
		r, e = processResponse(cmd, data)
		if r.Code == EX_OK {
			e = nil
		}
		return
	}, nil)
}

//external wrapper to simple call
func (s *Client) SimpleCall(cmd string, msgpars ...string) (reply *SpamDOut, err error) {
	return s.simpleCall(strings.ToUpper(cmd), msgpars)
}

//Tell what type of we are to process and what should be done
//with that message.  This includes setting or removing a local
//or a remote database (learning, reporting, forgetting, revoking)
func (s *Client) Tell(msgpars []string, headers *map[string]string) (reply *SpamDOut, err error) {
	return s.call(TELL, msgpars, func(data *bufio.Reader) (r *SpamDOut, e error) {
		r, e = processResponse(TELL, data)

		if r.Code == EX_UNAVAILABLE {
			e = errors.New("TELL commands are not enabled, set the --allow-tell switch.")
			return
		}
		if r.Code == EX_OK {
			e = nil
			return
		}
		return
	}, headers)
}

//here a TCP socket is created to call SPAMD
func (s *Client) call(cmd string, msgpars []string, onData FnCallback, extraHeaders *map[string]string) (reply *SpamDOut, err error) {

	if extraHeaders == nil {
		extraHeaders = &map[string]string{}
	}

	switch len(msgpars) {
	case 1:
		if s.User != "" {
			x := *extraHeaders
			x["User"] = s.User
			*extraHeaders = x
		}
	case 2:
		x := *extraHeaders
		x["User"] = msgpars[1]
		*extraHeaders = x
	default:
		if cmd != PING {
			err = errors.New("Message parameters wrong size")
		} else {
			msgpars = []string{""}
		}
		return
	}

	if cmd == REPORT_IGNOREWARNING {
		cmd = REPORT
	}

	// Create a new connection
	stream, err := net.Dial("tcp", s.Host)

	if err != nil {
		err = errors.New("Connection dial error to spamd: " + err.Error())
		return
	}
	// Set connection timeout
	timeout := time.Now().Add(time.Duration(s.ConnTimoutSecs) * time.Duration(time.Second))
	errTimeout := stream.SetDeadline(timeout)
	if errTimeout != nil {
		err = errors.New("Connection to spamd Timed Out:" + errTimeout.Error())
		return
	}
	defer stream.Close()

	// Create Command to Send to spamd
	cmd += " SPAMC/" + s.ProtocolVersion + "\r\n"
	cmd += "Content-length: " + fmt.Sprintf("%v\r\n", len(msgpars[0])+2)
	//Process Extra Headers if Any
	if len(*extraHeaders) > 0 {
		for hname, hvalue := range *extraHeaders {
			cmd = cmd + hname + ": " + hvalue + "\r\n"
		}
	}
	cmd += "\r\n" + msgpars[0] + "\r\n\r\n"

	_, errwrite := stream.Write([]byte(cmd))
	if errwrite != nil {
		err = errors.New("spamd returned a error: " + errwrite.Error())
		return
	}

	// Execute onData callback throwing the buffer like parameter
	reply, err = onData(bufio.NewReader(stream))
	return
}

//SpamD reply processor
func processResponse(cmd string, data *bufio.Reader) (returnObj *SpamDOut, err error) {
	defer func() {
		data.UnreadByte()
	}()

	returnObj = new(SpamDOut)
	returnObj.Code = -1
	//read the first line
	line, _, _ := data.ReadLine()
	lineStr := string(line)

	r := regexp.MustCompile(`(?i)SPAMD\/([0-9\.]+)\s([0-9]+)\s([0-9A-Z_]+)`)
	var result = r.FindStringSubmatch(lineStr)
	if len(result) < 4 {
		if cmd != "SKIP" {
			err = errors.New("spamd unreconized reply:" + lineStr)
		} else {
			returnObj.Code = EX_OK
			returnObj.Message = "SKIPPED"
		}
		return
	}
	returnObj.Code, _ = strconv.Atoi(result[2])
	returnObj.Message = result[3]

	//verify a mapped error...
	if SpamDError[returnObj.Code] != "" {
		err = errors.New(SpamDError[returnObj.Code])
		returnObj.Vars = make(map[string]interface{})
		returnObj.Vars["error_description"] = SpamDError[returnObj.Code]
		return
	}
	returnObj.Vars = make(map[string]interface{})

	//start didSet
	if cmd == TELL {
		returnObj.Vars["didSet"] = false
		returnObj.Vars["didRemove"] = false
		for {
			line, _, err = data.ReadLine()

			if err == io.EOF || err != nil {
				if err == io.EOF {
					err = nil
				}
				break
			}
			if strings.Contains(string(line), "DidRemove") {
				returnObj.Vars["didRemove"] = true
			}
			if strings.Contains(string(line), "DidSet") {
				returnObj.Vars["didSet"] = true
			}

		}
		return
	}
	//read the second line
	line, _, err = data.ReadLine()

	//finish here if line is empty
	if len(line) == 0 {
		if err == io.EOF {
			err = nil
		}
		return
	}

	//ignore content-length header..
	lineStr = string(line)
	switch cmd {

	case SYMBOLS, CHECK, REPORT, REPORT_IFSPAM, REPORT_IGNOREWARNING, PROCESS, HEADERS:

		switch cmd {
		case SYMBOLS, REPORT, REPORT_IFSPAM, REPORT_IGNOREWARNING, PROCESS, HEADERS:
			//ignore content-length header..
			line, _, err = data.ReadLine()
			lineStr = string(line)
		}

		r := regexp.MustCompile(`(?i)Spam:\s(True|False|Yes|No)\s;\s([0-9\.]+)\s\/\s([0-9\.]+)`)
		var result = r.FindStringSubmatch(lineStr)

		if len(result) > 0 {
			returnObj.Vars["isSpam"] = false
			switch result[1][0:1] {
			case "T", "t", "Y", "y":
				returnObj.Vars["isSpam"] = true
			}
			returnObj.Vars["spamScore"], _ = strconv.ParseFloat(result[2], 64)
			returnObj.Vars["baseSpamScore"], _ = strconv.ParseFloat(result[3], 64)
		}

		switch cmd {
		case PROCESS, HEADERS:
			lines := ""
			for {
				line, _, err = data.ReadLine()
				if err == io.EOF || err != nil {
					if err == io.EOF {
						err = nil
					}
					return
				}
				lines += string(line) + "\r\n"
				returnObj.Vars["body"] = lines
			}
			return
		case SYMBOLS:
			//ignore line break...
			data.ReadLine()
			//read
			line, _, err = data.ReadLine()
			returnObj.Vars["symbolList"] = strings.Split(string(line), ",")

		case REPORT, REPORT_IFSPAM, REPORT_IGNOREWARNING:
			//ignore line break...
			data.ReadLine()

			for {
				line, _, err = data.ReadLine()

				if len(line) > 0 {
					lineStr = string(line)

					//TXT Table found, prepare to parse..
					if len(lineStr) >= 4 && lineStr[0:4] == TABLE_MARK {

						section := []map[string]interface{}{}
						tt := 0
						for {
							line, _, err = data.ReadLine()
							//Stop read the text table if last line or Void line
							if err == io.EOF || err != nil || len(line) == 0 {
								if err == io.EOF {
									err = nil
								}
								break
							}
							//Parsing
							lineStr = string(line)
							spc := 2
							if lineStr[0:1] == "-" {
								spc = 1
							}
							lineStr = strings.Replace(lineStr, " ", SPLIT, spc)
							lineStr = strings.Replace(lineStr, " ", SPLIT, 1)
							if spc > 1 {
								lineStr = " " + lineStr[2:]
							}
							x := strings.Split(lineStr, SPLIT)
							if lineStr[1:3] == SPLIT {
								section[tt-1]["message"] = fmt.Sprintf("%v %v", section[tt-1]["message"], strings.TrimSpace(lineStr[5:]))
							} else {
								if len(x) != 0 {
									message := strings.TrimSpace(x[2])
									score, _ := strconv.ParseFloat(strings.TrimSpace(x[0]), 64)

									section = append(section, map[string]interface{}{
										"score":   score,
										"symbol":  x[1],
										"message": message,
									})

									tt++
								}
							}
						}
						if REPORT_IGNOREWARNING == cmd {
							nsection := []map[string]interface{}{}
							for _, c := range section {
								if c["score"].(float64) != 0 {
									nsection = append(nsection, c)
								}
							}
							section = nsection
						}

						returnObj.Vars["report"] = section
						break
					}
				}

				if err == io.EOF || err != nil {
					if err == io.EOF {
						err = nil
					}
					break
				}
			}
		}
	}

	if err != io.EOF {
		for {
			line, _, err = data.ReadLine()
			if err == io.EOF || err != nil {
				if err == io.EOF {
					err = nil
				}
				break
			}
		}
	}
	return
}
