(defproject simulator "0.1.0-SNAPSHOT"
  :dependencies [[org.clojure/clojure "1.10.0"]
                 [lispyclouds/clj-docker-client "0.2.2"]
                 [cheshire "5.8.1"]
                 [oc-humanize "0.2.3-alpha1"]]
  :main ^:skip-aot simulator.core
  :target-path "target/%s"
  :profiles {:uberjar {:aot :all}})
