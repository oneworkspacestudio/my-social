import{defineConfig}from'@playwright/test';
export default defineConfig({testDir:'tests',timeout:120000,workers:1,use:{baseURL:process.env.MYSOCIAL_TEST_URL||'http://localhost:29113',browserName:'chromium',channel:'chrome',headless:true,launchOptions:{args:['--use-fake-ui-for-media-stream','--use-fake-device-for-media-stream']},viewport:{width:1440,height:1000}},reporter:[['list']]});
