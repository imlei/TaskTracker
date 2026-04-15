(function () {
  var KEY = "pw_onboard_draft";

  window.pwOnboard = {
    load: function () {
      try {
        var raw = sessionStorage.getItem(KEY);
        return raw ? JSON.parse(raw) : {};
      } catch (e) {
        return {};
      }
    },
    save: function (partial) {
      var d = Object.assign(this.load(), partial);
      sessionStorage.setItem(KEY, JSON.stringify(d));
      return d;
    },
    clear: function () {
      sessionStorage.removeItem(KEY);
    },
  };
})();
