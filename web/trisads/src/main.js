"use strict";

var moment = require('moment');
var proto = require('./pb/api_grpc_web_pb');

var views = {}
views.alert = require('./views/alert.mustache');
views.lookupResult = require('./views/lookupResult.mustache');
views.searchResults = require('./views/searchResults.mustache');
views.otherJurisdiction = require('./views/otherJurisdiction.mustache');

(function ($) {
  $.fn.serializeForm = function() {
    var o = {};
    var a = this.serializeArray();
    $.each(a, function() {
      if (o[this.name]) {
        if (!o[this.name].push) {
          o[this.name] = [o[this.name]];
        }
        o[this.name].push(this.value || '');
      } else {
        o[this.name] = this.value || '';
      }
    })
    return o;
  };
})($);

const serializeForm = elements => [].reduce.call(elements, (data, element) => {
  data[element.name] = element.value;
  return data;
}, {});


const alert = (cls, err, msg) => {
  var elem = views.alert({ class: cls, error: err, message: msg })
  $("#alerts").html(elem);
  setTimeout(() => { $(".alert").alert("close"); }, 2000);
}

// TRISA Directory Client
var client = null

$(document).ready(function() {
  // Connect to the TRISA directory client and make queries.
  if (process.env.TRISADS_API_ENDPOINT) {
    client = new proto.TRISADirectoryClient(process.env.TRISADS_API_ENDPOINT);
    console.log("accessing TRISA directory at " + process.env.TRISADS_API_ENDPOINT);
  } else {
    client = new proto.TRISADirectoryClient('http://127.0.0.1:8080');
    console.log("accessing TRISA directory at http://127.0.0.1:8000");
  }

  // Bind the search form to the search action
  $("#searchForm").submit((e) => {
    e.preventDefault();
    var data = $(e.target).serializeForm();
    var req = new proto.SearchRequest();

    if (data.name) {
      var names = data.name.split(",");
      names = names.map(s=>s.trim());
      req.setNameList(names);
    }

    if (data.country) {
      var countries = data.country.split(",");
      countries = countries.map(s => s.trim());
      req.setCountryList(countries);
    }

    client.search(req, {}, (err, response) => {
      if (err || !response) {
        console.log(err);
        alert("danger", "connection error:", "no response from directory service");
        return
      }

      var err = response.getError();
      if (err) {
        alert("warning", "search error:", err.getMessage());
        return
      }

      var data = [];
      response.getVaspsList().forEach(v => {
        var entity = v.getVaspentity();
        data.push({
          id: v.getId(),
          name: entity.getVaspfulllegalname(),
          url: entity.getVaspurl(),
          category: entity.getVaspcategory(),
          country: entity.getVaspcountry()
        });
      });

      $("#searchResults").html(views.searchResults({results: data}));
    })

    return false;
  });

  // Bind the lookup form to the lookup action
  $("#lookupForm").submit((e) => {
    e.preventDefault();
    var data = $(e.target).serializeForm();
    var req = new proto.LookupRequest();

    if (data.lookupByID) {
      var id = parseInt(data.query);
      if (isNaN(id)) {
        var msg = "could not parse '" + data.query + "' into an integer ID";
        alert("danger", "invalid input", msg);
        return
      }
      req.setId(id);
    } else {
      req.setName(data.query);
    }

    client.lookup(req, {}, (err, response) => {
      if (err || !response) {
        console.log(err);
        alert("danger", "connection error:", "no response from directory service");
        return
      }

      var err = response.getError();
      if (err) {
        alert("warning", "could not lookup VASP:", err.getMessage());
        return
      }

      var vasp = response.getVasp();
      data = {
        "firstListed": moment(vasp.getFirstlisted()).format('MMMM Do YYYY'),
        "lastUpdated": moment(vasp.getLastupdated()).fromNow()
      }

      var entity = vasp.getVaspentity();
      if (entity) {
        data.entity = {
          "name": entity.getVaspfulllegalname(),
          "address": entity.getVaspfulllegaladdress(),
          "url": entity.getVaspurl(),
          "country": entity.getVaspcountry()
        };
      }

      var cert = vasp.getVasptrisacertification();
      if (cert) {
        data.cert = {
          "subject": cert.getSubjectname(),
          "issuer": cert.getIssuername(),
          "version": cert.getVersion(),
          "revoked": cert.getRevoked(),
          "keyInfo": cert.getPublickeyinfo().getSignature()
        };
      } else {
        data.cert = null;
      }

      $("#lookupResults").html(views.lookupResult(data));
    });

    return false
  });

  // Bind the register form to the register action
  $("#registerForm").submit((e) => {
    e.preventDefault();
    if (e.target.checkValidity() === false) {
      event.stopPropagation();
      e.target.classList.add('was-validated');
      return false
    }

    var data = $(e.target).serializeForm();
    console.log(data);
    alert("danger", "not yet implemented:", "currently this is only an example form");
    return false
  });

  // Add another jurisdiction to the register form
  $("#addOtherJurisdiction").click((e) => {
    $("#otherJurisdictions").append(views.otherJurisdiction());
  })

  // Only allow KYC threshold to be available if yes on question 3b.
  $('input[name="conductsCDD"').change((e) => {
    var ans = $(e.target).val();
    $("#kycThreshold").prop("disabled", ans=="no");
  });

  // Only allow Applicable Regulations to be available if yes on question 3d.
  $('input[name="mustComplyTravelRule"').change((e) => {
    var ans = $(e.target).val();
    $("#travelRuleRegulations").prop("disabled", ans == "no");
  });

});

