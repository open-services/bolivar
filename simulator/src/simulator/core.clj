(ns simulator.core
  (:require [clj-docker-client.core :as docker]
            [clojure.test :refer [is]]
            [cheshire.core :refer [parse-string]]
            [clojure.contrib.humanize :refer [filesize]])
  (:gen-class))

;; (comment
;;   ;; start five registries with share off
;;   ;; download packages
;;   ;; start five registries with share on
;;   ;; download packages
;;   ;; second version is faster
;; )
;; 
;; ;; faster as it's using sharing
;; (def open-registry-id (run-open-registry))
;; (def bolivar-ids (map (partial run bolivar) (range 5)))
;; 
;; ;; todo
;; ;; - [ ] image for open-registry
;; ;; - [ ] rename open-registry-fed -> bolivar
;; ;; - [ ] methods for
;; ;;  - [ ] running open-registry
;; ;;  - [ ] running bolivar
;; ;;  - [ ] running commands in existing container

(def open-registry {:image "open-services/open-registry:latest"})

(def bolivar {:image "open-services/bolivar:latest"
              :cmd "bolivar --http-address=0.0.0.0"})

(def deps-test {:image "open-services/deps-test:latest"})

(def conn (docker/connect))

(defn run [app]
  (clojure.pprint/pprint app)
  (docker/start conn (docker/create conn (:image app) (:cmd app) {} (:ports app))))

(defn exec [container-id cmd]
  (println container-id)
  (println cmd))

;; Stolen from
;; https://github.com/clojure/clojure/blob/clojure-1.9.0/src/clj/clojure/core.clj#L3850
(defmacro measure
  "Evaluates expr and return the time (in milliseconds) it took together with the result"
  [expr]
  `(let [start# (. System (nanoTime))
         ret# ~expr
         duration# (/ (double (- (. System (nanoTime)) start#)) 1000000.0)]
     {:return ret#
      :duration duration#}))

(comment
  (measure (+ 1 1))
  (measure (Thread/sleep 1000))
  )

(defn run-tests [tests]
  (doseq [t tests]
    (let [ret (t)]
      (clojure.pprint/pprint ret))))

(defn get-ip [container-id]
  (-> (docker/inspect conn container-id)
      :NetworkSettings
      :Networks
      :bridge
      :IPAddress
      ))

(defn run-with-registry [ip]
  (format "bash -c \"yarn --verbose --registry=http://%s:8080\"" ip))

(defn deps-test-image [ip]
  {:image "open-services/deps-test:latest"
   :cmd (run-with-registry ip)})

(defn run-and-wait [conn image]
  (let [id (run image)]
    (docker/wait-container conn id)
    id))

(defn measure-install [conn bolivar-ip]
  (measure (run-and-wait conn (deps-test-image bolivar-ip))))

(defn bolivar-metrics [bolivar-ip]
  (slurp (format "http://%s:8080/_api/metrics" bolivar-ip)))

(defn clean [ids]
  (doseq [id ids]
    (try (docker/kill conn id)
         (catch Exception e (println (str "didnt find container " id))))
    (docker/rm conn id)))

(defn get-metric [bolivar-ip k]
  (k (parse-string (bolivar-metrics bolivar-ip) true)))

(defn humanize-metrics [bolivar-ip]
  (let [m(parse-string (bolivar-metrics bolivar-ip) true)]
  {:TotalIn (filesize (:TotalIn m))
   :TotalOut (filesize (:TotalOut m))
   :RateIn (filesize (:RateIn m))
   :RateOut (filesize (:RateOut m))}))

;; (clojure.pprint/pprint (humanize-metrics bolivar-ip))

;; test cases

(defn second-install-cached []
  (let [bolivar-id (run bolivar)
        bolivar-ip (get-ip bolivar-id)]
    (Thread/sleep 1000)
    (let [m1 (measure-install conn bolivar-ip)
          m2 (measure-install conn bolivar-ip)]
      (is (< (:duration m2) (:duration m1)))
      (clean [(:return m2) (:return m1)
              bolivar-id]))))

(defn second-bolivar-faster []
  (let [bolivar-id (run bolivar)
        bolivar-ip (get-ip bolivar-id)
        bolivar2-id (run bolivar)
        bolivar2-ip (get-ip bolivar2-id)]
    (Thread/sleep 1000)
    (let [m1 (measure-install conn bolivar-ip)
          m2 (measure-install conn bolivar2-ip)]
      (let [ti1 (get-metric bolivar-ip :TotalIn)
            ti2 (get-metric bolivar2-ip :TotalIn)
            to1 (get-metric bolivar-ip :TotalOut)
            to2 (get-metric bolivar2-ip :TotalOut)]
        (is (< (:duration m2) (:duration m1)))
        (is (> to1 ti2))
        (is (> to1 to2))
        (is (> ti1 0))
        (is (> ti2 0))
        (is (> to1 0))
        (is (> to2 0))
        (clean [(:return m2) (:return m1)
                bolivar-id bolivar2-id])))))

(defn -main
  [& args]
  (run-tests [second-install-cached
              second-bolivar-faster])
  (System/exit 0))
