package scraper

import (
	"fmt"
	"net/http"
  "io/ioutil"
  "time"
  "regexp"
  "strconv"
  "bytes"
	"errors"
  "github.com/neocortical/oaklogger/searcher"
)

type ScrapedMessage struct {
  ThreadName string
  UID int
  Username string
  Message string
  MessageTime time.Time
  Edited bool
}

const timeFormat = "15:04PM on 02 Jan 2006"

func Scrape() {
	threadSearcher := searcher.NewThreadSearcher()
	postSearcher := searcher.NewPostSearcher()
	userSearcher := searcher.NewUserSearcher()
	defer threadSearcher.Close()
	defer postSearcher.Close()
	defer userSearcher.Close()
	
	pid := postSearcher.GetCurrentStatus() + 1
	consecutiveFails := 0
	fmt.Printf("Scraper process starting at PID %d\n", pid)
	
  for {
		m, err := ScrapeMessage(fmt.Sprintf("http://talk.oaklog.com/index.php?view=%d", pid))
		
		if (err != nil) {
			fmt.Printf("Error: %v\n", err)
			postSearcher.UpdateStatus(pid, "fail")
			consecutiveFails++
			
			if consecutiveFails == 25 {
				fmt.Printf("Detected end of messages, sleeping...\n")
				time.Sleep(15 * time.Minute)
				pid = postSearcher.GetCurrentStatus() + 1
				consecutiveFails = 0
				fmt.Printf("...waking up and retrying messages from PID %d\n", pid)
			}
		} else {
			thread := threadSearcher.FindThread(m.ThreadName)
			if thread == nil {
				fmt.Printf("Inserting new thread: %s\n", m.ThreadName)
				thread = buildThreadFromMessage(pid, m)
			} else {
				thread.PostCount++
				thread.LastPostTime = m.MessageTime
			}
			threadSearcher.Save(thread)
			
			postSearcher.Save(buildPostFromMessage(pid, m, thread))
			postSearcher.UpdateStatus(pid, "success")
			
			user := userSearcher.FindUser(m.UID)
			if user == nil {
				userSearcher.Save(buildUserFromMessage(m))
			}
			
			consecutiveFails = 0
		}
		
		fmt.Printf("Processed pid: %d\n", pid)
		time.Sleep(100 * time.Millisecond)
		pid++
	} 
}

func ScrapeMessage(url string) (*ScrapedMessage, error) {
	resp, err := http.Get(url)
  if err != nil {
  	return nil, err
  }
  defer resp.Body.Close()
  body, err := ioutil.ReadAll(resp.Body)
	if (err != nil) {
		return nil, err
	}
  
	return ParseBody(body)
}

func ParseBody(body []byte) (*ScrapedMessage, error) {
  re := regexp.MustCompile("<title>\\s*((?ms:.*))\\s*\\-\\s+talk\\.oaklog\\.com</title>")
  reThread := re.FindSubmatchIndex(body)
  
  re = regexp.MustCompile("<a href=http://www\\.oaklog\\.com/\\?a=view&u=([0-9]+) class=list><b>([^<]+)</b></a>")
  reUser := re.FindSubmatchIndex(body)
	
	if len(reUser) == 0 {
		return nil, errors.New("unable to parse message")
	}
	
  uid, _ := strconv.Atoi(string(body[reUser[2]:reUser[3]]))

  re = regexp.MustCompile("<td style=\"width: 85%; border\\-left: 0px solid #FFFFFF; border\\-top: 0px solid #FFFFFF\">\\s*((?ms:.+?))\\s*</td>")
  reMessage := re.FindSubmatchIndex(body)

  re = regexp.MustCompile("(Last edited on\\s+)?([0-9]{2}):([0-9]{2})(am|pm) on ([0-9]{2}) (Jan|Feb|Mar|Apr|May|Jun|Jul|Aug|Sep|Oct|Nov|Dec) ([0-9]{4})")
  reTime := re.FindSubmatchIndex(body)
  time := parseMessageTime(body, reTime)
    
  return &ScrapedMessage{
    string(body[reThread[2]:reThread[3]]), 
    uid, 
    string(body[reUser[4]:reUser[5]]), 
    string(body[reMessage[2]:reMessage[3]]), 
    time, 
    reTime[2] != -1}, nil
}

func parseMessageTime(body []byte, i []int) time.Time {
  hour, _ := strconv.Atoi(string(body[i[4]:i[5]]))
  minute, _ := strconv.Atoi(string(body[i[6]:i[7]]))
  if bytes.Compare(body[i[8]:i[9]], []byte("pm")) == 0 {
    if hour < 12 {
      hour += 12
    }
  } else if hour == 12 {
    hour = 0
  }
  
  day, _ := strconv.Atoi(string(body[i[10]:i[11]]))
  month := time.January
  
  switch string(body[i[12]:i[13]]) {
  case "Jan":
    month = time.January
  case "Feb":
    month = time.February
  case "Mar":
    month = time.March
  case "Apr":
    month = time.April
  case "May":
    month = time.May
  case "Jun":
    month = time.June
  case "Jul":
    month = time.July
  case "Aug":
    month = time.August
  case "Sep":
    month = time.September
  case "Oct":
    month = time.October
  case "Nov":
    month = time.November
  case "Dec":
    month = time.December
  }
  
  year, _ := strconv.Atoi(string(body[i[14]:i[15]]))
  
  return time.Date(year, month, day, hour, minute, 0, 0, time.Local)
}

func buildThreadFromMessage(pid int, m *ScrapedMessage) (*searcher.Thread) {
	result := new(searcher.Thread)
	result.TID  = pid
	result.Name = m.ThreadName
	result.UID = m.UID
	result.PostCount = 1
	result.LastPostTime = m.MessageTime
	return result
}

func buildPostFromMessage(pid int, m *ScrapedMessage, t *searcher.Thread) (*searcher.Post) {
	result := new(searcher.Post)
	result.PID = pid
	result.TID = t.TID
	result.UID = m.UID
	result.Order = t.PostCount
	result.Message = m.Message
	result.PostTime = m.MessageTime
	return result
}

func buildUserFromMessage(m *ScrapedMessage) (*searcher.User) {
	result := new(searcher.User)
	result.UID = m.UID
	result.Username = m.Username
	return result
}
