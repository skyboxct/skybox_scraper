import sys
import gspread
import time
import requests
from bs4 import BeautifulSoup
from oauth2client.service_account import ServiceAccountCredentials

# TODO: Optimize Speed
#

def main():
    print("Starting Scraper")

    headers = {'User-Agent': 'Mozilla/5.0 (Macintosh; Intel Mac OS X 10.15; rv:76.0) Gecko/20100101 Firefox/76.0'}
    scope = ['https://spreadsheets.google.com/feeds', 'https://www.googleapis.com/auth/drive']
    creds = ServiceAccountCredentials.from_json_keyfile_name('GoogleCreds.json', scope)
    client = gspread.authorize(creds)
    spreadsheet = client.open('Square Sports Scraper Ver2.0')
    sheet = spreadsheet.get_worksheet(1)


    rows = sheet.get_all_records()
    #Get all of the URLs in one object
    da_url_list = sheet.batch_get(['N2:N'+str(len(rows) + 1)])[0]
    sc_url_list = sheet.batch_get(['O2:O' + str(len(rows) + 1)])[0]

    row = 2
    for start_index in range(0, len(da_url_list), 10):

        da_batch_list = []


        for url in da_url_list[start_index:start_index+10]: #The first item is a list containing all the information
            da_range = "A" + str(row) + ":F" + str(row)

            if url:
                print("Reading url info for ", "line:",row, url[0])

                try:
                    response = requests.get(url[0], headers)
                except requests.exceptions.MissingSchema:
                    return

                soup = BeautifulSoup(response.content, features="lxml")

                da_title = soup.find('h1').get_text().strip()
                try:
                    #Grab price info
                    da_price = float(soup.find(class_='price large').get_text().strip("$").replace(',', ''))
                except AttributeError:
                    da_price = ""


                da_desc = soup.find(class_='eight columns').get_text().strip() #grab description
                da_pic = soup.find(class_='panel radius').findChild()['href'] #grab pic info
                da_upc = "UPC/Barcode" in soup.find(id='itemdetailsTab').find_all("li")[3].get_text()
                da_upc_text = ""
                da_stock_text = ""

                #logic for barcode information
                if da_upc:
                    da_upc_text = soup.find(id='itemdetailsTab').find_all("li")[3].get_text().strip("UPC/Barcode:")

                #logic for stock information
                if da_price == "":
                    da_stock_text = "Out Of Stock"
                else:
                    da_stock_text = "In Stock"


                da_batch_list.append({'range': da_range, 'values': [[da_title, da_price, da_desc, da_pic, da_upc_text, da_stock_text]]})
            row += 1
        da_total_list = sheet.batch_update(da_batch_list)

    row = 2
    for start_index in range(0, len(sc_url_list), 10):

        sc_batch_list = []


        for url in sc_url_list[start_index:start_index+10]: #The first item is a list containing all the information
            sc_range = "G" + str(row) + ":J" + str(row)

            if url:
                print("Reading url info for ", "line:",row, url[0])

                try:
                    response = requests.get(url[0], headers)
                except requests.exceptions.MissingSchema:
                    return

                soup = BeautifulSoup(response.content, features="lxml")

                sc_title = soup.find(itemprop='name').get_text().strip()
                try:
                    sc_price = float(soup.find(itemprop='price').get_text().strip().replace(',', ''))
                except AttributeError:
                    sc_price = ""

                sc_pic = soup.find('source').find_next()['srcset']
                sc_stock = "Sorry, but this item is currently out of stock." in soup.find(class_="five columns p-buy-box").get_text()
                sc_stock_text = ""

                # logic for stock information
                if sc_stock:
                    sc_price = ""
                    sc_stock_text = "Out of Stock"
                else:
                    sc_stock_text = "In Stock"

                sc_batch_list.append({'range': sc_range, 'values': [[sc_title, sc_price, sc_pic, sc_stock_text]]})
            row += 1
        sc_total_info = sheet.batch_update(sc_batch_list)







if __name__ == "__main__":
    main()

#   "client_email": "amazon-scrape@amazon-data-scrape.iam.gserviceaccount.com",

