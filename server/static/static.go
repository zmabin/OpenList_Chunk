package static

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"strings"

	"github.com/OpenListTeam/OpenList/v4/drivers/base"
	"github.com/OpenListTeam/OpenList/v4/internal/conf"
	"github.com/OpenListTeam/OpenList/v4/internal/setting"
	"github.com/OpenListTeam/OpenList/v4/pkg/utils"
	"github.com/OpenListTeam/OpenList/v4/public"
	"github.com/gin-gonic/gin"
)

type ManifestIcon struct {
	Src   string `json:"src"`
	Sizes string `json:"sizes"`
	Type  string `json:"type"`
}

type Manifest struct {
	Display  string         `json:"display"`
	Scope    string         `json:"scope"`
	StartURL string         `json:"start_url"`
	Name     string         `json:"name"`
	Icons    []ManifestIcon `json:"icons"`
}

var static fs.FS

func initStatic() {
	utils.Log.Debug("Initializing static file system...")
	if conf.Conf.DistDir == "" {
		dist, err := fs.Sub(public.Public, "dist")
		if err != nil {
			utils.Log.Fatalf("failed to read dist dir: %v", err)
		}
		static = dist
		utils.Log.Debug("Using embedded dist directory")
		return
	}
	static = os.DirFS(conf.Conf.DistDir)
	utils.Log.Infof("Using custom dist directory: %s", conf.Conf.DistDir)
}

// loginScheduleSidebarScript is injected into ManageHtml to add a
// "Login Schedule" link in the admin sidebar as a sub-item under Tasks.
const loginScheduleSidebarScript = `<script>
(function(){
  var T={zh_CN:"定时登录",zh_TW:"定時登入",ja:"ログインスケジュール",ko:"로그인 스케줄"};
  var lang=navigator.language.replace("-","_");
  var text=T[lang]||T[lang.split("_")[0]]||"Login Schedule";
  var base=location.pathname.replace(/\/@manage.*/,"");
  var href=base+"/@manage/login-schedule";
  var icon='<svg viewBox="0 0 512 512" width="16" height="16" fill="currentColor" style="flex-shrink:0;margin-right:8px"><path d="M256 8C119 8 8 119 8 256s111 248 248 248 248-111 248-248S393 8 258 8zm0 448c-110.5 0-200-89.5-200-200S145.5 56 256 56s200 89.5 200 200-89.5 200-200 200zm61.8-104.4l-84.9-61.7c-3.1-2.3-4.9-5.9-4.9-9.7V116c0-6.6 5.4-12 12-12h10c6.6 0 12 5.4 12 12v141.4l72.9 53.2c5.4 3.9 6.5 11.4 2.6 16.8l-8.2 11.3c-3.9 5.4-11.4 6.5-16.8 2.6z"/></svg>';
  function inject(){
    if(document.querySelector("[data-fork-login-schedule]")) return false;
    /* Find the Tasks <a> element (exact path, no trailing slash/children) */
    var links=document.querySelectorAll("a[href]");
    var tasks=null;
    for(var i=0;i<links.length;i++){
      var h=links[i].getAttribute("href");
      if(h&&h.indexOf("/@manage/tasks")!==-1&&h.indexOf("/@manage/tasks/")===-1){tasks=links[i];break;}
    }
    if(!tasks) return false;
    /* Style the link to match existing sub-items */
    var cs=getComputedStyle(tasks);
    var a=document.createElement("a");
    a.setAttribute("data-fork-login-schedule","1");
    a.href=href;
    a.innerHTML=icon+'<span style="flex:1">'+text+'</span>';
    a.style.cssText="display:flex;align-items:center;padding:"+cs.padding+";padding-left:2em;border-radius:"+cs.borderRadius+";font-size:"+cs.fontSize+";color:inherit;text-decoration:none;transition:background .15s";
    a.onmouseenter=function(){this.style.backgroundColor="rgba(0,0,0,.05)"};
    a.onmouseleave=function(){this.style.backgroundColor="transparent"};
    /* Navigate from <a> → Flex (parent) → Box (grandparent) */
    var flex=tasks.parentElement;
    if(!flex) return false;
    var box=flex.parentElement;
    if(!box) return false;
    /* Verify we're in a flex container (SideMenuItemWithChildren structure) */
    var fc=getComputedStyle(flex);
    if(fc.display!=="flex"){box=flex;flex=null;}
    /* Insert as sub-item: between the Tasks header (Flex) and the children list.
       Creates a wrapper div if needed to avoid being hidden with Show children. */
    if(flex&&box){
      var wrap=document.createElement("div");
      wrap.setAttribute("data-fork-login-schedule","1");
      wrap.style.cssText="padding:0 0 0 0";
      wrap.appendChild(a);
      box.insertBefore(wrap,flex.nextSibling);
    }else{
      var tgt=flex||box;
      tgt.parentElement.insertBefore(a,tgt.nextSibling);
    }
    return true;
  }
  if(inject()) return;
  var timer=setInterval(function(){if(inject())clearInterval(timer);},300);
  setTimeout(function(){clearInterval(timer);},10000);
  var obs=new MutationObserver(function(){inject()});
  obs.observe(document.getElementById("root")||document.body,{childList:true,subtree:true});
})();
</script>`

func replaceStrings(content string, replacements map[string]string) string {
	for old, new := range replacements {
		content = strings.Replace(content, old, new, 1)
	}
	return content
}

func initIndex(siteConfig SiteConfig) {
	utils.Log.Debug("Initializing index.html...")
	// dist_dir is empty and cdn is not empty, and web_version is empty or beta or dev or rolling
	if conf.Conf.DistDir == "" && conf.Conf.Cdn != "" && (conf.WebVersion == "" || conf.WebVersion == "beta" || conf.WebVersion == "dev" || conf.WebVersion == "rolling") {
		utils.Log.Infof("Fetching index.html from CDN: %s/index.html...", siteConfig.Cdn)
		resp, err := base.RestyClient.R().
			SetHeader("Accept", "text/html").
			Get(fmt.Sprintf("%s/index.html", siteConfig.Cdn))
		if err != nil {
			utils.Log.Fatalf("failed to fetch index.html from CDN: %v", err)
		}
		if resp.StatusCode() != http.StatusOK {
			utils.Log.Fatalf("failed to fetch index.html from CDN, status code: %d", resp.StatusCode())
		}
		conf.RawIndexHtml = string(resp.Body())
		utils.Log.Info("Successfully fetched index.html from CDN")
	} else {
		utils.Log.Debug("Reading index.html from static files system...")
		indexFile, err := static.Open("index.html")
		if err != nil {
			if errors.Is(err, fs.ErrNotExist) {
				utils.Log.Fatalf("index.html not exist, you may forget to put dist of frontend to public/dist")
			}
			utils.Log.Fatalf("failed to read index.html: %v", err)
		}
		defer func() {
			_ = indexFile.Close()
		}()
		index, err := io.ReadAll(indexFile)
		if err != nil {
			utils.Log.Fatalf("failed to read dist/index.html")
		}
		conf.RawIndexHtml = string(index)
		utils.Log.Debug("Successfully read index.html from static files system")
	}
	utils.Log.Debug("Replacing placeholders in index.html...")
	// Construct the correct manifest path based on basePath
	manifestPath := "/manifest.json"
	if siteConfig.BasePath != "/" {
		manifestPath = siteConfig.BasePath + "/manifest.json"
	}
	replaceMap := map[string]string{
		"cdn: undefined":                    fmt.Sprintf("cdn: '%s'", siteConfig.Cdn),
		"base_path: undefined":              fmt.Sprintf("base_path: '%s'", siteConfig.BasePath),
		`href="/manifest.json"`:             fmt.Sprintf(`href="%s"`, manifestPath),
	}
	conf.RawIndexHtml = replaceStrings(conf.RawIndexHtml, replaceMap)
	UpdateIndex()
}

func UpdateIndex() {
	utils.Log.Debug("Updating index.html with settings...")
	favicon := setting.GetStr(conf.Favicon)
	logo := strings.Split(setting.GetStr(conf.Logo), "\n")[0]
	title := setting.GetStr(conf.SiteTitle)
	customizeHead := setting.GetStr(conf.CustomizeHead)
	customizeBody := setting.GetStr(conf.CustomizeBody)
	mainColor := setting.GetStr(conf.MainColor)
	utils.Log.Debug("Applying replacements for default pages...")
	replaceMap1 := map[string]string{
		"https://res.oplist.org/logo/logo.svg": favicon,
		"https://res.oplist.org/logo/logo.png": logo,
		"Loading...":                           title,
		"main_color: undefined":                fmt.Sprintf("main_color: '%s'", mainColor),
	}
	conf.ManageHtml = replaceStrings(conf.RawIndexHtml, replaceMap1)
	// Inject fork sidebar link into ManageHtml (admin pages only)
	conf.ManageHtml = replaceStrings(conf.ManageHtml, map[string]string{
		"<!-- customize body -->": loginScheduleSidebarScript,
	})
	// IndexHtml (public pages) uses the user's customize_head and customize_body settings
	conf.IndexHtml = replaceStrings(conf.ManageHtml, map[string]string{
		"<!-- customize head -->":   customizeHead,
		loginScheduleSidebarScript: customizeBody,
	})
	utils.Log.Debug("Index.html update completed")
}

func ManifestJSON(c *gin.Context) {
	// Get site configuration to ensure consistent base path handling
	siteConfig := getSiteConfig()
	
	// Get site title from settings
	siteTitle := setting.GetStr(conf.SiteTitle)
	
	// Get logo from settings, use the first line (light theme logo)
	logoSetting := setting.GetStr(conf.Logo)
	logoUrl := strings.Split(logoSetting, "\n")[0]

	// Use base path from site config for consistency
	basePath := siteConfig.BasePath

	// Determine scope and start_url
	// PWA scope and start_url should always point to our application's base path
	// regardless of whether static resources come from CDN or local server
	scope := basePath
	startURL := basePath

	manifest := Manifest{
		Display:  "standalone",
		Scope:    scope,
		StartURL: startURL,
		Name:     siteTitle,
		Icons: []ManifestIcon{
			{
				Src:   logoUrl,
				Sizes: "512x512",
				Type:  "image/png",
			},
		},
	}

	c.Header("Content-Type", "application/json")
	c.Header("Cache-Control", "public, max-age=3600") // cache for 1 hour
	
	if err := json.NewEncoder(c.Writer).Encode(manifest); err != nil {
		utils.Log.Errorf("Failed to encode manifest.json: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate manifest"})
		return
	}
}

func Static(r *gin.RouterGroup, noRoute func(handlers ...gin.HandlerFunc)) {
	utils.Log.Debug("Setting up static routes...")
	siteConfig := getSiteConfig()
	initStatic()
	initIndex(siteConfig)
	folders := []string{"assets", "images", "streamer", "static"}
	
	if conf.Conf.Cdn == "" {
		utils.Log.Debug("Setting up static file serving...")
		r.Use(func(c *gin.Context) {
			for _, folder := range folders {
				if strings.HasPrefix(c.Request.RequestURI, fmt.Sprintf("/%s/", folder)) {
					c.Header("Cache-Control", "public, max-age=15552000")
				}
			}
		})
		for _, folder := range folders {
			sub, err := fs.Sub(static, folder)
			if err != nil {
				utils.Log.Fatalf("can't find folder: %s", folder)
			}
			utils.Log.Debugf("Setting up route for folder: %s", folder)
			r.StaticFS(fmt.Sprintf("/%s/", folder), http.FS(sub))
		}
	} else {
		// Ensure static file redirected to CDN
		for _, folder := range folders {
			r.GET(fmt.Sprintf("/%s/*filepath", folder), func(c *gin.Context) {
				filepath := c.Param("filepath")
				c.Redirect(http.StatusFound, fmt.Sprintf("%s/%s%s", siteConfig.Cdn, folder, filepath))
			})
		}
	}

	utils.Log.Debug("Setting up catch-all route...")
	noRoute(func(c *gin.Context) {
		if c.Request.Method != "GET" && c.Request.Method != "POST" {
			c.Status(405)
			return
		}
		c.Header("Content-Type", "text/html")
		c.Status(200)
		if strings.HasPrefix(c.Request.URL.Path, "/@manage") {
			_, _ = c.Writer.WriteString(conf.ManageHtml)
		} else {
			_, _ = c.Writer.WriteString(conf.IndexHtml)
		}
		c.Writer.Flush()
		c.Writer.WriteHeaderNow()
	})
}
