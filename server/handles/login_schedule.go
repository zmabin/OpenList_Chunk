package handles

import (
	"github.com/gin-gonic/gin"
)

func LoginSchedulePage(c *gin.Context) {
	c.Header("Content-Type", "text/html; charset=utf-8")
	c.String(200, loginScheduleHTML)
}

const loginScheduleHTML = `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>Login Schedule - OpenList</title>
<style>
  *{margin:0;padding:0;box-sizing:border-box}
  :root{
    --primary:#1890ff;
    --primary-hover:#40a9ff;
    --primary-bg:rgba(24,144,255,.08);
    --bg:#f0f2f5;
    --card:#fff;
    --border:#e8e8e8;
    --text:rgba(0,0,0,.85);
    --text-secondary:rgba(0,0,0,.45);
    --success:#52c41a;
    --success-bg:#f6ffed;
    --error:#ff4d4f;
    --error-bg:#fff2f0;
    --warning:#faad14;
    --warning-bg:#fffbe6;
    --disabled-bg:#f5f5f5;
    --shadow:0 1px 2px rgba(0,0,0,.03),0 1px 6px -1px rgba(0,0,0,.02),0 2px 4px rgba(0,0,0,.02);
    --radius:8px;
  }
  body{font-family:-apple-system,BlinkMacSystemFont,"Segoe UI",Roboto,"Helvetica Neue",Arial,"Noto Sans",sans-serif,"Apple Color Emoji","Segoe UI Emoji";background:var(--bg);color:var(--text);min-height:100vh;font-size:14px;line-height:1.5715}
  .topbar{background:var(--card);border-bottom:1px solid var(--border);padding:0 24px;height:56px;display:flex;align-items:center;justify-content:space-between;position:sticky;top:0;z-index:100}
  .topbar h1{font-size:16px;font-weight:600;display:flex;align-items:center;gap:8px}
  .topbar h1 .icon{color:var(--primary)}
  .topbar a{color:var(--text-secondary);text-decoration:none;font-size:14px;display:flex;align-items:center;gap:4px;transition:color .2s}
  .topbar a:hover{color:var(--primary)}
  .container{max-width:960px;margin:0 auto;padding:24px 16px}
  .card{background:var(--card);border-radius:var(--radius);box-shadow:var(--shadow);overflow:hidden}
  .card-header{padding:16px 24px;border-bottom:1px solid var(--border);display:flex;align-items:center;justify-content:space-between}
  .card-header h2{font-size:15px;font-weight:600}
  .card-header .count{color:var(--text-secondary);font-size:13px;font-weight:400}
  table{width:100%;border-collapse:collapse;font-size:14px}
  thead th{padding:12px 16px;text-align:left;font-weight:500;color:var(--text-secondary);background:var(--bg);border-bottom:1px solid var(--border);font-size:13px;white-space:nowrap}
  tbody td{padding:14px 16px;border-bottom:1px solid var(--border);transition:background .15s}
  tbody tr:hover td{background:var(--primary-bg)}
  tbody tr:last-child td{border-bottom:none}
  .mount-path{font-family:SFMono-Regular,Menlo,Monaco,Consolas,"Liberation Mono","Courier New",monospace;font-size:13px;color:var(--text)}
  .driver-name{font-weight:500}
  .badge{display:inline-flex;align-items:center;padding:1px 8px;border-radius:10px;font-size:12px;font-weight:500;line-height:20px}
  .badge-work{color:var(--success);background:var(--success-bg);border:1px solid #b7eb8f}
  .badge-error{color:var(--error);background:var(--error-bg);border:1px solid #ffa39e}
  .badge-disabled{color:var(--text-secondary);background:var(--disabled-bg);border:1px solid var(--border)}
  .interval-input{width:80px;padding:4px 8px;border:1px solid var(--border);border-radius:6px;font-size:14px;text-align:center;outline:none;transition:border-color .2s,box-shadow .2s;background:var(--card)}
  .interval-input:focus{border-color:var(--primary);box-shadow:0 0 0 2px var(--primary-bg)}
  .hint{color:var(--text-secondary);font-size:12px;margin-top:2px}
  .btn{display:inline-flex;align-items:center;justify-content:center;padding:4px 15px;border:1px solid var(--border);border-radius:6px;font-size:14px;font-weight:400;cursor:pointer;transition:all .2s;background:var(--card);color:var(--text);height:32px;white-space:nowrap}
  .btn:hover{color:var(--primary);border-color:var(--primary)}
  .btn-primary{background:var(--primary);border-color:var(--primary);color:#fff}
  .btn-primary:hover{background:var(--primary-hover);border-color:var(--primary-hover);color:#fff}
  .btn-primary:disabled{opacity:.5;cursor:not-allowed}
  .msg{padding:2px 0;font-size:12px;min-height:18px;margin-top:2px}
  .msg-ok{color:var(--success)}
  .msg-err{color:var(--error)}
  .empty{text-align:center;padding:48px 20px;color:var(--text-secondary)}
  .empty svg{margin-bottom:12px;opacity:.3}
  .login-hint{text-align:center;padding:80px 20px;color:var(--text-secondary)}
  .login-hint a{color:var(--primary);text-decoration:none}
  .login-hint a:hover{text-decoration:underline}
  .loading{text-align:center;padding:60px;color:var(--text-secondary)}
  .loading .spinner{display:inline-block;width:20px;height:20px;border:2px solid var(--border);border-top-color:var(--primary);border-radius:50%;animation:spin .6s linear infinite;margin-right:8px;vertical-align:middle}
  @keyframes spin{to{transform:rotate(360deg)}}
  .action-col{display:flex;align-items:center;gap:8px;flex-wrap:wrap}
  @media(max-width:640px){
    .container{padding:12px 8px}
    thead th,tbody td{padding:10px 12px}
    .card-header{padding:12px 16px}
    .hide-mobile{display:none}
  }
</style>
</head>
<body>
<div class="topbar">
  <h1><span class="icon"><svg width="18" height="18" viewBox="0 0 512 512" fill="currentColor"><path d="M256 8C119 8 8 119 8 256s111 248 248 248 248-111 248-248S393 8 258 8zm0 448c-110.5 0-200-89.5-200-200S145.5 56 256 56s200 89.5 200 200-89.5 200-200 200zm61.8-104.4l-84.9-61.7c-3.1-2.3-4.9-5.9-4.9-9.7V116c0-6.6 5.4-12 12-12h10c6.6 0 12 5.4 12 12v141.4l72.9 53.2c5.4 3.9 6.5 11.4 2.6 16.8l-8.2 11.3c-3.9 5.4-11.4 6.5-16.8 2.6z"/></svg></span><span id="page-title">Login Schedule</span></h1>
  <a id="back-link" href="/@manage"><svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M19 12H5"/><polyline points="12 19 5 12 12 5"/></svg><span id="back-text">Back to Admin</span></a>
</div>
<div class="container">
  <div id="content">
    <div class="loading"><span class="spinner"></span><span id="loading-text">Loading...</span></div>
  </div>
</div>
<script>
(function(){
  // i18n
  var L={
    zh_CN:{title:"定时登录",back:"返回管理后台",loading:"加载中...",noStorages:"暂无存储",mountPath:"挂载路径",driver:"驱动",status:"状态",interval:"登录间隔 (分钟)",action:"操作",save:"保存",saving:"保存中...",saved:"已保存",failed:"获取存储失败",saveFailed:"保存失败",error:"错误",disabled:"已禁用",hint:"0 = 禁用",count:"共 {n} 个存储",expired:"会话已过期，请重新登录",login:"登录",networkErr:"网络错误"},
    zh_TW:{title:"定時登入",back:"返回管理後台",loading:"載入中...",noStorages:"暫無存儲",mountPath:"掛載路徑",driver:"驅動",status:"狀態",interval:"登入間隔 (分鐘)",action:"操作",save:"儲存",saving:"儲存中...",saved:"已儲存",failed:"獲取存儲失敗",saveFailed:"儲存失敗",error:"錯誤",disabled:"已停用",hint:"0 = 停用",count:"共 {n} 個存儲",expired:"會話已過期，請重新登入",login:"登入",networkErr:"網路錯誤"},
    ja:{title:"ログインスケジュール",back:"管理画面へ戻る",loading:"読み込み中...",noStorages:"ストレージがありません",mountPath:"マウントパス",driver:"ドライバ",status:"ステータス",interval:"ログイン間隔 (分)",action:"操作",save:"保存",saving:"保存中...",saved:"保存しました",failed:"ストレージ取得に失敗",saveFailed:"保存に失敗",error:"エラー",disabled:"無効",hint:"0 = 無効",count:"{n} 件のストレージ",expired:"セッション切れです。再ログインしてください",login:"ログイン",networkErr:"ネットワークエラー"},
    ko:{title:"로그인 스케줄",back:"관리로 돌아가기",loading:"로딩 중...",noStorages:"스토리지 없음",mountPath:"마운트 경로",driver:"드라이버",status:"상태",interval:"로그인 간격 (분)",action:"작업",save:"저장",saving:"저장 중...",saved:"저장됨",failed:"스토리지 가져오기 실패",saveFailed:"저장 실패",error:"오류",disabled:"비활성화",hint:"0 = 비활성화",count:"스토리지 {n}개",expired:"세션이 만료되었습니다. 다시 로그인하세요",login:"로그인",networkErr:"네트워크 오류"}
  };
  var lang=navigator.language.replace("-","_");
  var T=L[lang]||L[lang.split("_")[0]]||L["en"]||{};
  function t(k,a){var s=T[k]||k;if(a)s=s.replace("{n}",a);return s;}

  // Apply translations
  document.documentElement.lang=lang.replace("_","-");
  document.title=t("title")+" - OpenList";
  document.getElementById("page-title").textContent=t("title");
  document.getElementById("back-text").textContent=t("back");
  document.getElementById("loading-text").textContent=t("loading");

  var TOKEN_KEY="openlist-token";
  var apiBase=window.location.pathname.replace(/\/@manage\/login-schedule.*/,"");

  function getToken(){
    var keys=[TOKEN_KEY,"token","alist-token","open_list_token"];
    for(var i=0;i<keys.length;i++){var v=localStorage.getItem(keys[i]);if(v)return v;}
    return "";
  }

  function api(method,path,body){
    var opts={method:method,headers:{"Authorization":getToken()}};
    if(body){opts.headers["Content-Type"]="application/json";opts.body=JSON.stringify(body);}
    return fetch(apiBase+"/api/admin"+path,opts).then(function(r){return r.json();});
  }

  function esc(s){var d=document.createElement("div");d.textContent=s;return d.innerHTML;}

  function statusBadge(s){
    if(s==="work")return '<span class="badge badge-work">OK</span>';
    if(s==="disabled")return '<span class="badge badge-disabled">'+esc(t("disabled"))+'</span>';
    if(s)return '<span class="badge badge-error" title="'+esc(s)+'">'+esc(s.substring(0,20))+'</span>';
    return '<span class="badge badge-disabled">-</span>';
  }

  function render(storages){
    if(!storages||storages.length===0){
      return '<div class="card"><div class="empty"><svg width="48" height="48" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5"><path d="M20 7l-8-4-8 4m16 0l-8 4m8-4v10l-8 4m0-10L4 7m8 4v10M4 7v10l8 4"/></svg><p>'+t("noStorages")+'</p></div></div>';
    }
    var h='<div class="card"><div class="card-header"><h2>'+t("title")+'</h2><span class="count">'+t("count",storages.length)+'</span></div>';
    h+='<div style="overflow-x:auto"><table><thead><tr>';
    h+='<th>'+t("mountPath")+'</th>';
    h+='<th>'+t("driver")+'</th>';
    h+='<th>'+t("status")+'</th>';
    h+='<th>'+t("interval")+'</th>';
    h+='<th>'+t("action")+'</th>';
    h+='</tr></thead><tbody>';
    storages.forEach(function(s){
      var val=s.login_interval||0;
      h+='<tr>';
      h+='<td class="mount-path">'+esc(s.mount_path)+'</td>';
      h+='<td class="driver-name">'+esc(s.driver)+'</td>';
      h+='<td>'+statusBadge(s.status)+'</td>';
      h+='<td><input class="interval-input" type="number" min="0" value="'+val+'" data-id="'+s.id+'" id="interval-'+s.id+'"/>';
      h+='<div class="hint">'+t("hint")+'</div></td>';
      h+='<td><div class="action-col"><button class="btn btn-primary" onclick="saveInterval('+s.id+')">'+t("save")+'</button>';
      h+='<span class="msg" id="msg-'+s.id+'"></span></div></td>';
      h+='</tr>';
    });
    h+='</tbody></table></div></div>';
    return h;
  }

  function showMsg(id,text,ok){
    var el=document.getElementById("msg-"+id);
    if(el){el.textContent=text;el.className="msg "+(ok?"msg-ok":"msg-err");if(text)setTimeout(function(){el.textContent="";},3000);}
  }

  window.saveInterval=function(id){
    var input=document.getElementById("interval-"+id);
    if(!input)return;
    var val=parseInt(input.value)||0;
    showMsg(id,t("saving"),false);
    api("GET","/storage/get?id="+id).then(function(res){
      if(res.code!==200){showMsg(id,res.message||t("failed"),false);return;}
      var storage=res.data;
      storage.login_interval=val;
      return api("POST","/storage/update",storage);
    }).then(function(res){
      if(!res)return;
      if(res.code===200){showMsg(id,t("saved"),true);}
      else{showMsg(id,res.message||t("saveFailed"),false);}
    }).catch(function(e){showMsg(id,t("error")+": "+e.message,false);});
  };

  // Fetch main_color and apply
  api("GET","/setting/get?key=main_color").then(function(res){
    if(res.code===200&&res.data&&res.data.value){
      var mc=res.data.value;
      document.documentElement.style.setProperty("--primary",mc);
      // Compute hover color (slightly lighter)
      document.documentElement.style.setProperty("--primary-hover",mc+"dd");
      document.documentElement.style.setProperty("--primary-bg",mc.replace(")",",.08)").replace("rgb","rgba").replace("#","")?mc+"14":"rgba(24,144,255,.08)");
    }
  }).catch(function(){});

  // Apply computed primary-bg with fallback
  function applyPrimaryBg(){
    var p=getComputedStyle(document.documentElement).getPropertyValue("--primary").trim();
    if(p&&p.startsWith("#")){
      document.documentElement.style.setProperty("--primary-bg",p+"14");
    }
  }

  function init(){
    document.getElementById("back-link").href=apiBase+"/@manage";
    var token=getToken();
    if(!token){
      document.getElementById("content").innerHTML='<div class="card"><div class="login-hint"><p>'+t("expired")+'. <a href="'+apiBase+'/@manage">'+t("login")+'</a></p></div></div>';
      return;
    }
    applyPrimaryBg();
    api("GET","/storage/list?page=1&per_page=100").then(function(res){
      if(res.code===200){
        document.getElementById("content").innerHTML=render(res.data.content);
      }else if(res.code===401||res.code===403){
        document.getElementById("content").innerHTML='<div class="card"><div class="login-hint"><p>'+t("expired")+'. <a href="'+apiBase+'/@manage">'+t("login")+'</a></p></div></div>';
      }else{
        document.getElementById("content").innerHTML='<div class="card"><div class="login-hint"><p>'+t("error")+': '+esc(res.message||"unknown")+'</p></div></div>';
      }
    }).catch(function(e){
      document.getElementById("content").innerHTML='<div class="card"><div class="login-hint"><p>'+t("networkErr")+': '+esc(e.message)+'</p></div></div>';
    });
  }

  init();
})();
</script>
</body>
</html>`
