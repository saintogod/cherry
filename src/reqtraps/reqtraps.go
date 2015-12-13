package reqtraps

import (
    "net"
    "../config"
    "../html"
    "../rawhttp"
    "strings"
)

type RequestTrapInterface interface {
    Handle(new_conn net.Conn, room_name, http_payload string, rooms *config.CherryRooms, preprocessor *html.Preprocessor)
}

type RequestTrapHandleFunc func(new_conn net.Conn, room_name, http_payload string, rooms *config.CherryRooms, preprocessor *html.Preprocessor)

func (h RequestTrapHandleFunc) Handle(new_conn net.Conn, room_name, http_payload string, rooms *config.CherryRooms, preprocessor *html.Preprocessor) {
    h(new_conn, room_name, http_payload, rooms, preprocessor)
}

type RequestTrap func() RequestTrapInterface

func BuildRequestTrap(handle RequestTrapHandleFunc) RequestTrap {
    return func () RequestTrapInterface {
        return RequestTrapHandleFunc(handle)
    }
}

func GetRequestTrap(http_payload string) RequestTrap {
    var http_method_part string
    var space_nr int = 0
    for _, h := range http_payload {
        if h == ' ' {
            space_nr++
        }
        if h == '\n' || h == '\r' || space_nr == 2 {
            break
        }
        http_method_part += string(h)
    }
    http_method_part += "$"
    if strings.HasPrefix(http_method_part, "GET /join$") {
        return BuildRequestTrap(GetJoin_Handle)
    }
    if strings.HasPrefix(http_method_part, "GET /brief$") {
        return BuildRequestTrap(GetBrief_Handle)
    }
    if strings.HasPrefix(http_method_part, "GET /top&") {
        return BuildRequestTrap(GetTop_Handle)
    }
    if strings.HasPrefix(http_method_part, "GET /banner&") {
        return BuildRequestTrap(GetBanner_Handle)
    }
    if strings.HasPrefix(http_method_part, "GET /body&") {
        return BuildRequestTrap(GetBody_Handle)
    }
    if strings.HasPrefix(http_method_part, "GET /exit&") {
        return BuildRequestTrap(GetExit_Handle)
    }
    if strings.HasPrefix(http_method_part, "POST /join$") {
        return BuildRequestTrap(PostJoin_Handle)
    }
    if strings.HasPrefix(http_method_part, "POST /banner&") {
        return BuildRequestTrap(PostBanner_Handle)
    }
    if strings.HasPrefix(http_method_part, "GET /find$") {
        return BuildRequestTrap(GetFind_Handle)
    }
    if strings.HasPrefix(http_method_part, "POST /find$") {
        return BuildRequestTrap(PostFind_Handle)
    }
    return BuildRequestTrap(BadAssError_Handle)
}

func GetFind_Handle(new_conn net.Conn, room_name, http_payload string, rooms *config.CherryRooms, preprocessor *html.Preprocessor) {
    var reply_buffer []byte
    reply_buffer = rawhttp.MakeReplyBuffer(preprocessor.ExpandData(room_name, rooms.GetFindBotTemplate(room_name)), 200, true)
    new_conn.Write(reply_buffer)
    new_conn.Close()
}

func PostFind_Handle(new_conn net.Conn, room_name, http_payload string, rooms *config.CherryRooms, preprocessor *html.Preprocessor) {
    var user_data map[string]string
    user_data = rawhttp.GetFieldsFromPost(http_payload)
    var reply_buffer []byte
    if _, posted := user_data["user"]; !posted {
        reply_buffer = rawhttp.MakeReplyBuffer(html.GetBadAssErrorData(), 404, true)
    } else {
        var result string
        result = preprocessor.ExpandData(room_name, rooms.GetFindResultsHeadTemplate(room_name))
        listing := rooms.GetFindResultsBodyTemplate(room_name)
        avail_rooms := rooms.GetRooms()
        user := strings.ToUpper(user_data["user"])
        if len(user) > 0 {
            for _, r := range avail_rooms {
                users := rooms.GetRoomUsers(r)
                preprocessor.SetDataValue("{{.find-result-users-total}}", rooms.GetUsersTotal(r))
                preprocessor.SetDataValue("{{.find-result-room-name}}", r)
                for _, u := range users {
                    if strings.HasPrefix(strings.ToUpper(u), user) {
                        preprocessor.SetDataValue("{{.find-result-user}}", u)
                        result += preprocessor.ExpandData(room_name, listing)
                    }
                }
            }
        }
        result += preprocessor.ExpandData(room_name, rooms.GetFindResultsTailTemplate(room_name))
        reply_buffer = rawhttp.MakeReplyBuffer(result, 200, true)
    }
    new_conn.Write(reply_buffer)
    new_conn.Close()
}

func GetJoin_Handle(new_conn net.Conn, room_name, http_payload string, rooms *config.CherryRooms, preprocessor *html.Preprocessor) {
    //  INFO(Santiago): The form for room joining was requested, so we will flush it to client.
    var reply_buffer []byte
    reply_buffer = rawhttp.MakeReplyBuffer(preprocessor.ExpandData(room_name, rooms.GetEntranceTemplate(room_name)), 200, true)
    new_conn.Write(reply_buffer)
    new_conn.Close()
}

func GetTop_Handle(new_conn net.Conn, room_name, http_payload string, rooms *config.CherryRooms, preprocessor *html.Preprocessor) {
    var user_data map[string]string
    user_data = rawhttp.GetFieldsFromGet(http_payload)
    var reply_buffer []byte
    if !rooms.IsValidUserRequest(room_name, user_data["user"], user_data["id"]) {
        reply_buffer = rawhttp.MakeReplyBuffer(html.GetBadAssErrorData(), 404, true)
    } else {
        reply_buffer = rawhttp.MakeReplyBuffer(preprocessor.ExpandData(room_name, rooms.GetTopTemplate(room_name)), 200, true)
    }
    new_conn.Write(reply_buffer)
    new_conn.Close()
}

func GetBanner_Handle(new_conn net.Conn, room_name, http_payload string, rooms *config.CherryRooms, preprocessor *html.Preprocessor) {
    var user_data map[string]string
    var reply_buffer []byte
    user_data = rawhttp.GetFieldsFromGet(http_payload)
    preprocessor.SetDataValue("{{.nickname}}", user_data["user"])
    preprocessor.SetDataValue("{{.session-id}}", user_data["id"])
    if !rooms.IsValidUserRequest(room_name, user_data["user"], user_data["id"]) {
        reply_buffer = rawhttp.MakeReplyBuffer(html.GetBadAssErrorData(), 404, true)
    } else {
        reply_buffer = rawhttp.MakeReplyBuffer(preprocessor.ExpandData(room_name, rooms.GetBannerTemplate(room_name)), 200, true)
    }
    new_conn.Write(reply_buffer)
    new_conn.Close()
}

func GetExit_Handle(new_conn net.Conn, room_name, http_payload string, rooms *config.CherryRooms, preprocessor *html.Preprocessor) {
    var user_data map[string]string
    var reply_buffer []byte
    user_data = rawhttp.GetFieldsFromGet(http_payload)
    if !rooms.IsValidUserRequest(room_name, user_data["user"], user_data["id"]) {
        reply_buffer = rawhttp.MakeReplyBuffer(html.GetBadAssErrorData(), 404, true)
    } else {
        preprocessor.SetDataValue("{{.nickname}}", user_data["user"])
        preprocessor.SetDataValue("{{.session-id}}", user_data["id"])
        reply_buffer = rawhttp.MakeReplyBuffer(preprocessor.ExpandData(room_name, rooms.GetExitTemplate(room_name)), 200, true)
    }
    rooms.EnqueueMessage(room_name, user_data["user"], "", "", "", "",  rooms.GetExitMessage(room_name), "")
    new_conn.Write(reply_buffer)
    rooms.RemoveUser(room_name, user_data["user"])
    new_conn.Close()
}

func PostJoin_Handle(new_conn net.Conn, room_name, http_payload string, rooms *config.CherryRooms, preprocessor *html.Preprocessor) {
    //  INFO(Santiago): Here, we need firstly parse the posted fields, check for "nickclash", if this is the case
    //                  flush the page informing it. Otherwise we add the user basic info and flush the room skeleton
    //                  [TOP/BODY/BANNER]. Then we finally close the connection.
    var user_data map[string]string
    var reply_buffer []byte
    user_data = rawhttp.GetFieldsFromPost(http_payload)
    if _, posted := user_data["user"]; !posted {
        new_conn.Close()
        return
    }
    if _, posted := user_data["color"]; !posted {
        new_conn.Close()
        return
    }
    preprocessor.SetDataValue("{{.nickname}}", user_data["user"])
    preprocessor.SetDataValue("{{.session-id}}", "0")
    if rooms.HasUser(room_name, user_data["user"]) || user_data["user"] == rooms.GetAllUsersAlias(room_name) {
        reply_buffer = rawhttp.MakeReplyBuffer(preprocessor.ExpandData(room_name, rooms.GetNickclashTemplate(room_name)), 200, true)
    } else {
        rooms.AddUser(room_name, user_data["user"], user_data["color"], true)
        preprocessor.SetDataValue("{{.session-id}}", rooms.GetSessionId(user_data["user"], room_name))
        reply_buffer = rawhttp.MakeReplyBuffer(preprocessor.ExpandData(room_name, rooms.GetSkeletonTemplate(room_name)), 200, true)
        rooms.EnqueueMessage(room_name, user_data["user"], "", "", "", "", rooms.GetJoinMessage(room_name), "")
    }
    new_conn.Write(reply_buffer)
    new_conn.Close()
}

func GetBrief_Handle(new_conn net.Conn, room_name, http_payload string, rooms *config.CherryRooms, preprocessor *html.Preprocessor) {
    var reply_buffer []byte
    reply_buffer = rawhttp.MakeReplyBuffer(preprocessor.ExpandData(room_name, rooms.GetBriefTemplate(room_name)), 200, true)
    new_conn.Write(reply_buffer)
    new_conn.Close()
}

func GetBody_Handle(new_conn net.Conn, room_name, http_payload string, rooms *config.CherryRooms, preprocessor *html.Preprocessor) {
    var user_data map[string]string
    user_data = rawhttp.GetFieldsFromGet(http_payload)
    var valid_user bool
    valid_user = rooms.IsValidUserRequest(room_name, user_data["user"], user_data["id"])
    var reply_buffer []byte
    if !valid_user {
        reply_buffer = rawhttp.MakeReplyBuffer(html.GetBadAssErrorData(), 404, true)
    } else {
        reply_buffer = rawhttp.MakeReplyBuffer(preprocessor.ExpandData(room_name, rooms.GetBodyTemplate(room_name)), 200, false)
    }
    new_conn.Write(reply_buffer)
    if valid_user {
        rooms.SetUserConnection(room_name, user_data["user"], new_conn)
    } else {
        new_conn.Close()
    }
}

func BadAssError_Handle(new_conn net.Conn, room_name, http_payload string, rooms *config.CherryRooms, preprocessor *html.Preprocessor) {
    new_conn.Write(rawhttp.MakeReplyBuffer(html.GetBadAssErrorData(), 404, true))
    new_conn.Close()
}

func PostBanner_Handle(new_conn net.Conn, room_name, http_payload string, rooms *config.CherryRooms, preprocessor *html.Preprocessor) {
    var user_data map[string]string
    var reply_buffer []byte
    var invalid_request bool = false
    user_data = rawhttp.GetFieldsFromPost(http_payload)
    if _ , has := user_data["user"]; !has {
        invalid_request = true
    } else if _, has := user_data["id"]; !has {
        invalid_request = true
    } else if _, has := user_data["action"]; !has {
        invalid_request = true
    } else if _, has := user_data["whoto"]; !has {
        invalid_request = true
    } else if _, has := user_data["sound"]; !has {
        invalid_request = true
    } else if  _, has := user_data["image"]; !has {
        invalid_request = true
    } else if _, has := user_data["says"]; !has {
        invalid_request = true
    }
    var restore_banner bool = true
    if invalid_request || !rooms.IsValidUserRequest(room_name, user_data["user"], user_data["id"]) {
        reply_buffer = rawhttp.MakeReplyBuffer(html.GetBadAssErrorData(), 404, true)
    } else if user_data["action"] == rooms.GetIgnoreAction(room_name) {
        if user_data["user"] != user_data["whoto"] && ! rooms.IsIgnored(user_data["user"], user_data["whoto"], room_name) {
            rooms.AddToIgnoreList(user_data["user"], user_data["whoto"], room_name)
            rooms.EnqueueMessage(room_name, user_data["user"], "", "", "", "", rooms.GetOnIgnoreMessage(room_name) + user_data["whoto"], "1")
            restore_banner = false
        }
    } else if user_data["action"] == rooms.GetDeIgnoreAction(room_name) {
        if rooms.IsIgnored(user_data["user"], user_data["whoto"], room_name) {
            rooms.DelFromIgnoreList(user_data["user"], user_data["whoto"], room_name)
            rooms.EnqueueMessage(room_name, user_data["user"], "", "", "", "", rooms.GetOnDeIgnoreMessage(room_name) + user_data["whoto"], "1")
            restore_banner = false
        }
    } else {
        var something_to_say bool =  (len(user_data["says"]) > 0 || len(user_data["image"]) > 0 || len(user_data["sound"]) > 0)
        if something_to_say {
            //  INFO(Santiago): Any further antiflood control would go from here.
            rooms.EnqueueMessage(room_name, user_data["user"], user_data["whoto"], user_data["action"], user_data["sound"], user_data["image"], user_data["says"], user_data["priv"])
        }
    }
    preprocessor.SetDataValue("{{.nickname}}", user_data["user"])
    preprocessor.SetDataValue("{{.session-id}}", user_data["id"])
    if user_data["priv"] == "1" {
        preprocessor.SetDataValue("{{.priv}}", "checked")
    }
    temp_banner := preprocessor.ExpandData(room_name, rooms.GetBannerTemplate(room_name))
    if restore_banner {
        temp_banner = strings.Replace(temp_banner,
                                      "<option value = \"" + user_data["whoto"] + "\">",
                                      "<option value = \"" + user_data["whoto"] + "\" selected>", -1)
        temp_banner = strings.Replace(temp_banner,
                                      "<option value = \"" + user_data["action"] + "\">",
                                      "<option value = \"" + user_data["action"] + "\" selected>", -1)
    }
    reply_buffer = rawhttp.MakeReplyBuffer(temp_banner, 200, true)
    new_conn.Write(reply_buffer)
    new_conn.Close()
}
