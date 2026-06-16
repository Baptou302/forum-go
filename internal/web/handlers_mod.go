package web

import "net/http"

func (a *App) handlePinThread(w http.ResponseWriter, r *http.Request) {
	id := idParam(r)
	if t, err := a.store.GetThread(id, 0); err == nil {
		a.store.SetPinned(id, !t.Pinned)
	}
	redirectBack(w, r, threadPath(id))
}

func (a *App) handleLockThread(w http.ResponseWriter, r *http.Request) {
	id := idParam(r)
	if t, err := a.store.GetThread(id, 0); err == nil {
		a.store.SetLocked(id, !t.Locked)
	}
	redirectBack(w, r, threadPath(id))
}

func (a *App) handleDeleteThread(w http.ResponseWriter, r *http.Request) {
	a.store.DeleteThread(idParam(r))
	setFlash(w, "success", "Sujet supprimé.")
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (a *App) handleDeletePost(w http.ResponseWriter, r *http.Request) {
	a.store.DeletePost(idParam(r))
	redirectBack(w, r, "/")
}
