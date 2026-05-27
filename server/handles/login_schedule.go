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
    --hope-colors-primary9:#1890ff;
    --hope-colors-primary10:#40a9ff;
    --hope-colors-primary11:#69c0ff;
    --hope-colors-loContrast:#ffffff;
    --hope-colors-neutral3:#f5f5f5;
    --hope-colors-neutral4:#e8e8e8;
    --hope-colors-neutral11:rgba(0,0,0,0.85);
    --hope-colors-neutral12:rgba(0,0,0,0.65);
    --hope-colors-neutral13:rgba(0,0,0,0.45);
    --hope-colors-success9:#52c41a;
    --hope-colors-error9:#ff4d4f;
    --hope-colors-warning9:#faad14;
    --hope-space-1:4px;
    --hope-space-2:8px;
    --hope-space-3:12px;
    --hope-space-4:16px;
    --hope-space-5:20px;
    --hope-space-6:24px;
    --hope-radii-md:8px;
    --hope-radii-lg:12px;
    --hope-shadows-md:0 2px 8px rgba(0,0,0,0.08);
    --hope-fontSizes-sm:12px;
    --hope-fontSizes-md:14px;
    --hope-fontSizes-lg:16px;
    --hope-fontSizes-xl:20px;
  }
  body{
    font-family:-apple-system,BlinkMacSystemFont,"Segoe UI",Roboto,"Helvetica Neue",Arial,"Noto Sans",sans-serif;
    background:var(--hope-colors-neutral3);
    color:var(--hope-colors-neutral11);
    min-height:100vh;
    font-size:var(--hope-fontSizes-md);
    line-height:1.6;
  }

  /* Page container */
  .page{
    padding:var(--hope-space-6);
    max-width:1200px;
    margin:0 auto;
  }

  /* Page header */
  .page-header{
    margin-bottom:var(--hope-space-6);
    display:flex;
    align-items:center;
    justify-content:space-between;
  }
  .page-header h1{
    font-size:var(--hope-fontSizes-xl);
    font-weight:600;
    display:flex;
    align-items:center;
    gap:var(--hope-space-2);
  }
  .page-header .icon{
    color:var(--hope-colors-primary9);
    width:24px;
    height:24px;
  }

  /* Card container */
  .card{
    background:var(--hope-colors-loContrast);
    border-radius:var(--hope-radii-lg);
    box-shadow:var(--hope-shadows-md);
    overflow:hidden;
  }

  /* Table header */
  .table-header{
    padding:var(--hope-space-4) var(--hope-space-5);
    border-bottom:1px solid var(--hope-colors-neutral4);
    display:flex;
    align-items:center;
    justify-content:space-between;
  }
  .table-header h2{
    font-size:var(--hope-fontSizes-lg);
    font-weight:500;
  }
  .table-header .count{
    color:var(--hope-colors-neutral13);
    font-size:var(--hope-fontSizes-sm);
  }

  /* Table */
  table{
    width:100%;
    border-collapse:collapse;
    font-size:var(--hope-fontSizes-md);
  }
  thead th{
    padding:var(--hope-space-3) var(--hope-space-5);
    text-align:left;
    font-weight:500;
    color:var(--hope-colors-neutral13);
    background:var(--hope-colors-neutral3);
    border-bottom:1px solid var(--hope-colors-neutral4);
    font-size:var(--hope-fontSizes-sm);
  }
  tbody td{
    padding:var(--hope-space-4) var(--hope-space-5);
    border-bottom:1px solid var(--hope-colors-neutral4);
    transition:background 0.15s;
  }
  tbody tr:hover td{
    background:rgba(24,144,255,0.04);
  }
  tbody tr:last-child td{
    border-bottom:none;
  }

  /* Mount path styling */
  .mount-path{
    font-family:"SF Mono",SFMono-Regular,Menlo,Mono,Consolas,"Liberation Mono","Courier New",monospace;
    font-size:var(--hope-fontSizes-sm);
    color:var(--hope-colors-primary9);
  }

  /* Driver name */
  .driver-name{
    font-weight:500;
  }

  /* Status badges */
  .badge{
    display:inline-flex;
    align-items:center;
    padding:2px var(--hope-space-2);
    border-radius:12px;
    font-size:var(--hope-fontSizes-sm);
    font-weight:500;
    line-height:20px;
  }
  .badge-success{
    color:var(--hope-colors-success9);
    background:rgba(82,196,26,0.1);
    border:1px solid rgba(82,196,26,0.3);
  }
  .badge-error{
    color:var(--hope-colors-error9);
    background:rgba(255,77,79,0.1);
    border:1px solid rgba(255,77,79,0.3);
  }
  .badge-disabled{
    color:var(--hope-colors-neutral13);
    background:var(--hope-colors-neutral3);
    border:1px solid var(--hope-colors-neutral4);
  }

  /* Input */
  .interval-input{
    width:80px;
    padding:var(--hope-space-1) var(--hope-space-2);
    border:1px solid var(--hope-colors-neutral4);
    border-radius:var(--hope-radii-md);
    font-size:var(--hope-fontSizes-md);
    text-align:center;
    outline:none;
    transition:border-color 0.2s, box-shadow 0.2s;
    background:var(--hope-colors-loContrast);
  }
  .interval-input:focus{
    border-color:var(--hope-colors-primary9);
    box-shadow:0 0 0 2px rgba(24,144,255,0.1);
  }

  /* Hint */
  .hint{
    color:var(--hope-colors-neutral13);
    font-size:var(--hope-fontSizes-sm);
    margin-top:var(--hope-space-1);
  }

  /* Buttons */
  .btn{
    display:inline-flex;
    align-items:center;
    justify-content:center;
    padding:0 var(--hope-space-4);
    border:1px solid var(--hope-colors-neutral4);
    border-radius:var(--hope-radii-md);
    font-size:var(--hope-fontSizes-md);
    font-weight:500;
    cursor:pointer;
    transition:all 0.2s;
    background:var(--hope-colors-loContrast);
    color:var(--hope-colors-neutral11);
    height:32px;
    white-space:nowrap;
  }
  .btn:hover{
    color:var(--hope-colors-primary9);
    border-color:var(--hope-colors-primary9);
  }
  .btn-primary{
    background:var(--hope-colors-primary9);
    border-color:var(--hope-colors-primary9);
    color:white;
  }
  .btn-primary:hover{
    background:var(--hope-colors-primary10);
    border-color:var(--hope-colors-primary10);
    color:white;
  }
  .btn-primary:disabled{
    opacity:0.5;
    cursor:not-allowed;
  }

  /* Action column */
  .action-col{
    display:flex;
    align-items:center;
    gap:var(--hope-space-2);
  }

  /* Messages */
  .msg{
    font-size:var(--hope-fontSizes-sm);
    min-height:20px;
  }
  .msg-ok{
    color:var(--hope-colors-success9);
  }
  .msg-err{
    color:var(--hope-colors-error9);
  }

  /* Empty state */
  .empty{
    text-align:center;
    padding:80px var(--hope-space-6);
    color:var(--hope-colors-neutral13);
  }
  .empty svg{
    margin-bottom:var(--hope-space-3);
    opacity:0.3;
    width:48px;
    height:48px;
  }

  /* Login hint */
  .login-hint{
    text-align:center;
    padding:80px var(--hope-space-6);
    color:var(--hope-colors-neutral13);
  }
  .login-hint a{
    color:var(--hope-colors-primary9);
    text-decoration:none;
    font-weight:500;
  }
  .login-hint a:hover{
    text-decoration:underline;
  }

  /* Loading state */
  .loading{
    text-align:center;
    padding:80px var(--hope-space-6);
    color:var(--hope-colors-neutral13);
  }
  .spinner{
    display:inline-block;
    width:20px;
    height:20px;
    border:2px solid var(--hope-colors-neutral4);
    border-top-color:var(--hope-colors-primary9);
    border-radius:50%;
    animation:spin 0.8s linear infinite;
    margin-right:var(--hope-space-2);
    vertical-align:middle;
  }
  @keyframes spin{
    to{transform:rotate(360deg)}
  }

  /* Mobile responsive */
  @media(max-width:768px){
    .page{
      padding:var(--hope-space-4);
    }
    .table-header{
      padding:var(--hope-space-3) var(--hope-space-4);
    }
    thead th,tbody td{
      padding:var(--hope-space-3) var(--hope-space-4);
    }
    .page-header{
      flex-direction:column;
      align-items:flex-start;
      gap:var(--hope-space-2);
    }
  }
</style>
</head>
<body>
<div class="page">
  <div class="page-header">
    <h1>
      <svg class="icon" viewBox="0 0 512 512" fill="currentColor">
        <path d="M256 8C119 8 8 119 8 256s111 248 248 248 248-111 248-248S393 8 258 8zm0 448c-110.5 0-200-89.5-200-200S145.5 56 256 56s200 89.5 200 200-89.5 200-200 200zm61.8-104.4l-84.9-61.7c-3.1-2.3-4.9-5.9-4.9-9.7V116c0-6.6 5.4-12 12-12h10c6.6 0 12 5.4 12 12v141.4l72.9 53.2c5.4 3.9 6.5 11.4 2.6 16.8l-8.2 11.3c-3.9 5.4-11.4 6.5-16.8 2.6z"/>
      </svg>
      <span id="page-title">Login Schedule</span>
    </h1>
  </div>

  <div id="content">
    <div class="loading">
      <span class="spinner"></span>
      <span id="loading-text">Loading...</span>
    </div>
  </div>
</div>

<script>
(function(){
  var L={
    zh_CN:{title:"定时登录",loading:"加载中...",noStorages:"暂无存储",mountPath:"挂载路径",driver:"驱动",status:"状态",interval:"登录间隔（分钟）",save:"保存",saving:"保存中...",saved:"已保存",failed:"获取存储失败",saveFailed:"保存失败",error:"错误",disabled:"已禁用",hint:"0 = 禁用",count:"共 {n} 个存储",expired:"会话已过期，请重新登录",login:"登录",networkErr:"网络错误"},
    zh_TW:{title:"定時登入",loading:"載入中...",noStorages:"暫無存儲",mountPath:"掛載路徑",driver:"驅動",status:"狀態",interval:"登入間隔（分鐘）",save:"儲存",saving:"儲存中...",saved:"已儲存",failed:"獲取存儲失敗",saveFailed:"儲存失敗",error:"錯誤",disabled:"已停用",hint:"0 = 停用",count:"共 {n} 個存儲",expired:"會話已過期，請重新登入",login:"登入",networkErr:"網路錯誤"},
    ja:{title:"ログインスケジュール",loading:"読み込み中...",noStorages:"ストレージがありません",mountPath:"マウントパス",driver:"ドライバ",status:"ステータス",interval:"ログイン間隔（分）",save:"保存",saving:"保存中...",saved:"保存しました",failed:"ストレージ取得に失敗",saveFailed:"保存に失敗",error:"エラー",disabled:"無効",hint:"0 = 無効",count:"{n} 件のストレージ",expired:"セッション切れです。再ログインしてください",login:"ログイン",networkErr:"ネットワークエラー"},
    ko:{title:"로그인 스케줄",loading:"로딩 중...",noStorages:"스토리지 없음",mountPath:"마운트 경로",driver:"드라이버",status:"상태",interval:"로그인 간격（분）",save:"저장",saving:"저장 중...",saved:"저장됨",failed:"스토리지 가져오기 실패",saveFailed:"저장 실패",error:"오류",disabled:"비활성화",hint:"0 = 비활성화",count:"스토리지 {n}개",expired:"세션이 만료되었습니다. 다시 로그인하세요",login:"로그인",networkErr:"네트워크 오류"}
  };
  var lang=navigator.language.replace("-","_");
  var T=L[lang]||L[lang.split("_")[0]]||L["en"]||{};
  function t(k,a){var s=T[k]||k;if(a)s=s.replace("{n}",a);return s;}

  document.documentElement.lang=lang.replace("_","-");
  document.title=t("title")+" - OpenList";
  document.getElementById("page-title").textContent=t("title");
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
    if(s==="work")return '<span class="badge badge-success">'+t("ok")+'</span>';
    if(s==="disabled")return '<span class="badge badge-disabled">'+t("disabled")+'</span>';
    if(s)return '<span class="badge badge-error" title="'+esc(s)+'">'+esc(s.substring(0,20))+'</span>';
    return '<span class="badge badge-disabled">-</span>';
  }

  function render(storages){
    if(!storages||storages.length===0){
      return '<div class="card"><div class="empty"><svg viewBox="0 0 512 512" fill="currentColor"><path d="M192 32c0-17.7 14.3-32 32-32h64c17.7 0 32 14.3 32 32v48H192V32zm-32 48H96c-35.3 0-64 28.7-64 64v288c0 35.3 28.7 64 64 64h256c35.3 0 64-28.7 64-64V144c0-35.3-28.7-64-64-64h-64v-32c0-17.7-14.3-32-32-32h-64c-17.7 0-32 14.3-32 32v48zM416 288H352v64h64v-64z"/></svg><p>'+t("noStorages")+'</p></div></div>';
    }
    var h='<div class="card"><div class="table-header"><h2>'+t("title")+'</h2><span class="count">'+t("count",storages.length)+'</span></div>';
    h+='<table><thead><tr>';
    h+='<th>'+t("mountPath")+'</th>';
    h+='<th>'+t("driver")+'</th>';
    h+='<th>'+t("status")+'</th>';
    h+='<th>'+t("interval")+'</th>';
    h+='<th></th>';
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
    h+='</tbody></table></div>';
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

  // Fetch main_color and apply to CSS variables
  api("GET","/setting/get?key=main_color").then(function(res){
    if(res.code===200&&res.data&&res.data.value){
      var mc=res.data.value;
      document.documentElement.style.setProperty("--hope-colors-primary9",mc);
      // Compute lighter variants
      document.documentElement.style.setProperty("--hope-colors-primary10",mc+"dd");
    }
  }).catch(function(){});

  function init(){
    var token=getToken();
    if(!token){
      document.getElementById("content").innerHTML='<div class="card"><div class="login-hint"><p>'+t("expired")+'</p><p style="margin-top:8px"><a href="'+apiBase+'/@manage">'+t("login")+'</a></p></div></div>';
      return;
    }
    api("GET","/storage/list?page=1&per_page=100").then(function(res){
      if(res.code===200){
        document.getElementById("content").innerHTML=render(res.data.content);
      }else if(res.code===401||res.code===403){
        document.getElementById("content").innerHTML='<div class="card"><div class="login-hint"><p>'+t("expired")+'</p><p style="margin-top:8px"><a href="'+apiBase+'/@manage">'+t("login")+'</a></p></div></div>';
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
