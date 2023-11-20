<!-- markdownlint-configure-file {
  "MD013": false,
  "MD033": false
} -->

<h1 align="center">
  Grafana Infinity Datasource
</h1>

<p align="center">Visualize data from JSON, CSV, XML, GraphQL and HTML endpoints in Grafana.</p>

<p align="center">
  <a href="#-key-features">Key Features</a> •
  <a href="#%EF%B8%8F-download">Download</a> •
  <a href="#%EF%B8%8F-documentation">Documentation</a> •
  <a href="#%EF%B8%8F-useful-links">Useful links</a> •
  <a href="#%EF%B8%8F-project-assistance">Project assistance</a> •
  <a href="#%EF%B8%8F-license">License</a>
</p>

<p align="center">
    <a href="https://yesoreyeram.github.io/grafana-infinity-datasource">
      <img src="https://raw.githubusercontent.com/yesoreyeram/grafana-infinity-datasource/main/src/img/icon.svg" alt="Grafana Infinity Datasource" width=140">
    </a>
</p>


## 👍 Contributing ONLY 

This is a forked REPOSITORY. It is not signed or published publicly. 
The only way to see changes is to follow the contributing guide.
- Git Clone this repository and make the necessary commands within the plugin directory to build and run.
- My steps (in powershell/command lines):
  1. Run npm i
  2. yarn or yarn isntall
  3. yarn dev (ctr + c ,exits out of continuous code checking)
  4. mage -v (backend build)
  5. docker-compose up
  6. Go to the site localhost:3001 (credentials-> admin:apofapof)
  7. In the side bar, navigate to Administration->Plugins
  8. Type in Infinity (it should say unsigned)
  9. Create a Infinity data source and configure it
  10. In the Authentication tab you should see the ZCAP tab with a resource target field. There is where your edv url or whatever url you will try to download and request will go.
  11. Save and Test.
  12. Navigate back to the Datasources page and Build a dashboard with your new configured Infinity Plugin. (That is where the mercury-client ERROR occurs)


- Read the [contributing guide](https://github.com/yesoreyeram/grafana-infinity-datasource/blob/main/CONTRIBUTING.md) for more details

## ⭐️ Project assistance

If you want to say **thank you** or/and support active development of `Grafana Infinity Datasource`:

- Add a [GitHub Star](https://github.com/yesoreyeram/grafana-infinity-datasource) to the project.
- Tweet about project [on your Twitter](https://twitter.com/intent/tweet?text=Checkout%20this%20cool%20%23grafana%20datasource%20%40grafanainfinity.%20%0A%0ALiterally,%20get%20your%20data%20from%20anywhere%20into%20%23grafana.%20JSON,%20CSV,%20XML,%20GraphQL,%20OAuth2,%20RSS%20feed,%20%23kubernetes,%20%23azure,%20%23aws,%20%23gcp%20and%20more%20stuff.%0A%0Ahttps%3A//yesoreyeram.github.io/grafana-infinity-datasource%0A).
- Write articles about project on [Dev.to](https://dev.to/), [Medium](https://medium.com/) or personal blog.

## ⚠️ License

This project is licensed under [Apache 2.0](https://github.com/yesoreyeram/grafana-infinity-datasource/blob/main/LICENSE)
