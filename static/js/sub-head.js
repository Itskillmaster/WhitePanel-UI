/* تم قبل از اولین رندر اعمال می‌شود: پیش‌فرض روشن، بدون فلش شب→روز هنگام رفرش */
(function(){
  try{
    if (localStorage.getItem("white-theme-v2") === "dark"){
      document.documentElement.setAttribute("data-theme","dark");
      var m = document.querySelector('meta[name="theme-color"]');
      if (m) m.setAttribute("content","#01091a");
    }
  }catch(e){}
})();
