GopherTown.Router.map(function() {
  this.resource('index', { path: '/' });
});
GopherTown.Router.map(function() {
  this.resource('index', { path: '/index.html' });
});

GopherTown.Router.map(function() {
  this.route('search', { path: 'search/:needle' });
});

GopherTown.Router.map(function() {
  this.resource('random', { path: '/random' });
});

GopherTown.Router.map(function() {
  this.resource('people', { path: '/:username' });
});

GopherTown.PeopleRoute = Ember.Route.extend({
  model: function(params) {
    return jQuery.getJSON('/gophers/user?username=' + params.username).then(function(res) {
      return { results: res };
    });
  },

  // TODO
  //serialize: function(model) {
  //  return { username: model.get('username') };
  //}
});


GopherTown.SearchRoute = Ember.Route.extend({
  model: function(params) {
    return jQuery.getJSON('/gophers/search?for=' + params.needle).then(function(res) {
      return { results: res };
    });
  }
});


GopherTown.ApplicationRoute = Ember.Route.extend({
  actions: {
    search: function(val) {
      this.transitionTo('search', val);
    }
  }
});


GopherTown.RandomRoute = Ember.Route.extend({
  model: function(params) {
    return jQuery.getJSON('/gophers/random').then(function(res) {
        console.log(res)
      return { results: res };
    });
  }
});
