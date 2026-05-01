package utils
import(
	"strings"
)

func CategorizeReferrer(ref string) string {
    if ref == "" || ref == "Direct / Bookmark" {
        return "Directo / Apps"
    }

    ref = strings.ToLower(ref)

    switch {
    case strings.Contains(ref, "facebook.com") || strings.Contains(ref, "fb.com"):
        return "Facebook"
    case strings.Contains(ref, "t.co") || strings.Contains(ref, "twitter.com") || strings.Contains(ref, "x.com"):
        return "X / Twitter"
    case strings.Contains(ref, "instagram.com"):
        return "Instagram"
    case strings.Contains(ref, "linkedin.com"):
        return "LinkedIn"
    case strings.Contains(ref, "youtube.com") || strings.Contains(ref, "youtu.be"):
        return "YouTube"
    case strings.Contains(ref, "google."):
        return "Google Search"
    default:
        // Si es otro sitio, extraemos solo el dominio (ej: blog.com)
        parts := strings.Split(ref, "/")
        if len(parts) > 2 {
            return parts[2] 
        }
        return "Otros"
    }
}