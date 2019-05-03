(defproject simulator "0.1.0-SNAPSHOT"
  :dependencies [[org.clojure/clojure "1.10.0"]
                 [lispyclouds/clj-docker-client "0.2.2"]]
  :main ^:skip-aot simulator.core
  :target-path "target/%s"
  :profiles {:uberjar {:aot :all}})
