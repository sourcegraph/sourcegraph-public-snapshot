// +build js

package main

import (
	"encoding/json"
	"log"
	"strconv"

	"github.com/gopherjs/gopherjs/js"
	"honnef.co/go/js/dom"

	"src.sourcegraph.com/apps/tracker/issues"
)

// State for tracking current reactions on the page.
var commentReactions map[uint64][]issues.Reaction

func OpenReactionMenu(el dom.Element, commentID uint64) {
	rect := el.GetBoundingClientRect()

	reactionMenu := dom.GetWindow().Document().GetElementByID("reaction-menu")

	var clearMenu *js.Object
	opt := js.M{
		"x": rect.Left,
		"y": rect.Top + 20,
		"onClose": func() {
			clearMenu.Call("call")
		},
		"onSelect": func(name string) {
			ToggleReaction(commentID, issues.EmojiID(name))
		},
	}
	clearMenu = js.Global.Get("Sourcegraph").Get("Components").Call("emojiMenu", reactionMenu, opt)
}

func RenderReactionList(el dom.Element, commentID uint64, reactions []issues.Reaction) {
	opt := js.M{
		"reactions": reactions,
	}
	if state.CurrentUser != nil {
		opt["onSelect"] = func(name string) {
			ToggleReaction(commentID, issues.EmojiID(name))
		}
	} else {
		opt["onSelect"] = func(string) {
			// Since the user isn't logged in, display the reaction menu which will prompt the user to login.
			OpenReactionMenu(el, commentID)
		}
	}
	js.Global.Get("Sourcegraph").Get("Components").Call("reactionList", el, opt)
}

func ToggleReaction(commentID uint64, emojiID issues.EmojiID) {
	applyReactionToggle(commentID, emojiID)
	reactionLists := dom.GetWindow().Document().GetElementsByClassName("comment-reactions-container")
	RenderReactionList(reactionLists[commentID], commentID, commentReactions[commentID])

	go func() {
		err := toggleReaction(strconv.FormatUint(commentID, 10), emojiID)
		if err != nil {
			// TODO: Handle failure properly.
			log.Println(err)
		}
	}()
}

// applyReactionToggle updates the local state that tracks reactions for
// comments on the current page.
func applyReactionToggle(commentID uint64, emojiID issues.EmojiID) {

	cr := commentReactions[commentID]
	for i := range cr {
		if cr[i].Reaction == emojiID {
			// Toggle this user's reaction.
			switch reacted := containsUser(cr[i].Users, *state.CurrentUser); {
			case reacted == -1:
				// Add this reaction.
				cr[i].Users = append(cr[i].Users, *state.CurrentUser)
			case reacted >= 0:
				// Remove this reaction.
				cr[i].Users[reacted] = cr[i].Users[len(cr[i].Users)-1] // Delete without preserving order.
				cr[i].Users = cr[i].Users[:len(cr[i].Users)-1]

				// If there are no more authors backing it, this reaction goes away.
				if len(cr[i].Users) == 0 {
					// TODO: Use single-line "delete preserving order" style, after
					// https://github.com/gopherjs/gopherjs/issues/358 is resolved.
					cr = append(cr[:i], cr[i+1:]...)
				}
			}
			commentReactions[commentID] = cr
			return
		}
	}
	// If we get here, this is the first reaction of its kind.
	// Add it to the end of the list.
	cr = append(cr,
		issues.Reaction{
			Reaction: emojiID,
			Users:    []issues.User{*state.CurrentUser},
		},
	)
	commentReactions[commentID] = cr
}

func setupReactions() {
	commentReactions = make(map[uint64][]issues.Reaction)

	reactionLists := dom.GetWindow().Document().GetElementsByClassName("comment-reactions-container")
	for _, el := range reactionLists {
		var rs []issues.Reaction
		json.Unmarshal([]byte(el.GetAttribute("data-reactions")), &rs)
		cid, _ := strconv.ParseUint(el.GetAttribute("data-id"), 10, 64)
		commentReactions[cid] = rs
		RenderReactionList(el, cid, rs)
	}
}

func containsUser(users []issues.User, user issues.User) int {
	for i, u := range users {
		if u.ID == user.ID {
			return i
		}
	}
	return -1
}
