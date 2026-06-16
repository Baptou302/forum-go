package web

import (
	"fmt"
	"html/template"
	"strings"
	"time"
)

func initialOf(s string) string {
	if s == "" {
		return "?"
	}
	return strings.ToUpper(string([]rune(s)[:1]))
}

func dict(values ...any) map[string]any {
	m := map[string]any{}
	for i := 0; i+1 < len(values); i += 2 {
		m[fmt.Sprint(values[i])] = values[i+1]
	}
	return m
}

func timeAgo(t time.Time) string {
	d := time.Since(t)
	switch {
	case d < time.Minute:
		return "à l'instant"
	case d < time.Hour:
		n := int(d.Minutes())
		return fmt.Sprintf("il y a %d %s", n, plural(n, "minute", "minutes"))
	case d < 24*time.Hour:
		n := int(d.Hours())
		return fmt.Sprintf("il y a %d %s", n, plural(n, "heure", "heures"))
	case d < 30*24*time.Hour:
		n := int(d.Hours() / 24)
		return fmt.Sprintf("il y a %d %s", n, plural(n, "jour", "jours"))
	default:
		return "le " + dateFR(t)
	}
}

var moisFR = []string{"janv.", "févr.", "mars", "avr.", "mai", "juin",
	"juil.", "août", "sept.", "oct.", "nov.", "déc."}

func dateFR(t time.Time) string {
	return fmt.Sprintf("%d %s %d", t.Day(), moisFR[int(t.Month())-1], t.Year())
}

func nl2br(s string) template.HTML {
	return template.HTML(strings.ReplaceAll(template.HTMLEscapeString(s), "\n", "<br>"))
}

func excerpt(s string, n int) string {
	s = strings.TrimSpace(strings.ReplaceAll(s, "\n", " "))
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	return strings.TrimSpace(string(r[:n])) + "…"
}

func plural(n int, singular, pluriel string) string {
	if n > 1 || n < -1 {
		return pluriel
	}
	return singular
}

func roleLabel(role string) string {
	switch role {
	case "admin":
		return "Administrateur"
	case "moderator":
		return "Modérateur"
	default:
		return "Membre"
	}
}

func roleBadge(role string) string {
	switch role {
	case "admin":
		return "bg-red-100 text-red-700 dark:bg-red-500/15 dark:text-red-300"
	case "moderator":
		return "bg-amber-100 text-amber-700 dark:bg-amber-500/15 dark:text-amber-300"
	default:
		return "bg-slate-100 text-slate-600 dark:bg-slate-700/50 dark:text-slate-300"
	}
}
