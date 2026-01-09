package server

// asyncAuditor will try to make merged commads out of keystrokes
func asyncAuditor() {
	SysLogger.Debug().Msg("starting asyncAuditor")
	for {
		audit, ok := <-asyncAuditChan
		if !ok {
			SysLogger.Debug().Msg("channel closed, stopping asyncAuditor")
			break
		}
		storeOrFlush(audit)
	}
}

// storeOrFlush will push keystrokes into a byte slice and
// flush it upen enter or a certain limit
func storeOrFlush(audit asyncAudit) {
	for _, ascii := range audit.ascii {
		switch ascii {
		case 0:
			// nothing
		case 8, 127:
			commandSync.Lock()
			if len(commandMap[audit.ctxid]) > 0 {
				commandMap[audit.ctxid] = commandMap[audit.ctxid][:len(commandMap[audit.ctxid])-1]
			}
			commandSync.Unlock()
		case 13:
			commandSync.Lock()
			command := string(commandMap[audit.ctxid])
			commandMap[audit.ctxid] = nil
			commandSync.Unlock()
			
			mapSync.Lock()
			logCommand(command, userMap[audit.ctxid], audit.ctxid, namespaceMap[audit.ctxid], podMap[audit.ctxid], containerMap[audit.ctxid], remoteAddrMap[audit.ctxid])
			mapSync.Unlock()
		default:
			commandSync.Lock()
			// to prevent oom kills by shoving too much input into one line
			// we flush after the amount of strokes set in MaxStokesPerLine
			if len(commandMap[audit.ctxid]) > MaxStokesPerLine {
				command := string(commandMap[audit.ctxid])
				commandMap[audit.ctxid] = nil
				commandSync.Unlock()
				
				mapSync.Lock()
				logCommand(command, userMap[audit.ctxid], audit.ctxid, namespaceMap[audit.ctxid], podMap[audit.ctxid], containerMap[audit.ctxid], remoteAddrMap[audit.ctxid])
				mapSync.Unlock()
				
				commandSync.Lock()
			}
			commandMap[audit.ctxid] = append(commandMap[audit.ctxid], ascii)
			commandSync.Unlock()
		}

	}
}

type asyncAudit struct {
	ctxid string
	ascii []byte
}
