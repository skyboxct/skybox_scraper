import gspread
import pyppeteer
import requests_html
from bs4 import BeautifulSoup
from oauth2client.service_account import ServiceAccountCredentials
from requests_html import HTMLSession

MAX_RETRIES = 5

da_output_range = ["A", "F"]
tcg_output_range = ["G", "K"]
tnt_output_range = ["L", "O"]
cs_output_range = ["P", "S"]

# "client_email": "tcgscraper@tcg-scraper.iam.gserviceaccount.com"
# TODO: Optimize Speed


def main():
    print("Starting Scraper")

    scope = ['https://spreadsheets.google.com/feeds', 'https://www.googleapis.com/auth/drive']
    creds = ServiceAccountCredentials.from_json_keyfile_name('tcgcreds.json', scope)
    client = gspread.authorize(creds)
    spreadsheet = client.open('Square TCG Scraper Ver1.1')
    sheet = spreadsheet.get_worksheet(1)
    records = sheet.get_all_records()

    scrape_products(records, sheet)


def scrape_products(records, sheet):
    # Dictionary to hold host configurations
    host_configs = {
        "DA": {"range": da_output_range, "parser_func": parse_da},
        "TCG": {"range": tcg_output_range, "parser_func": parse_tcg},
        "TNT": {"range": tnt_output_range, "parser_func": parse_tnt},
        "CS": {"range": cs_output_range, "parser_func": parse_cs}
    }

    row = 2
    for start_index in range(0, len(records), 10):
        batch_list = []
        for product in records[start_index:start_index + 10]:
            print("\nReading product information for row", row)
            urls = [product["DA url"], product["TCG url"], product["TNT url"], product["CS url"]]
            product_pages = get_product_pages(urls)
            for page in product_pages:
                print("   Reading page for", page["url"])
                host = page["host"]
                page_html = page["html"]

                if host == "TCG":
                    parser_parameter = page_html
                else:
                    parser_parameter = BeautifulSoup(page_html.html.raw_html, features="lxml")

                output_range = host_configs[host]["range"][0] + str(row) + ":" + host_configs[host]["range"][1] + str(row)
                values = retry(host_configs[host]["parser_func"], parser_parameter, MAX_RETRIES)

                if output_range and values:
                    print("   Values: ", values)
                    batch_list.append({'range': output_range, 'values': values})
            row += 1
        print("\nPushing lines", start_index + 2, "-", start_index + 12, "to sheet\n")
        sheet.batch_update(batch_list)
    # TODO: roll up parsing errors into output before exit


def parse_da(soup):
    title = soup.find('h1').get_text().strip()
    try:
        # Grab price info
        da_price = float(soup.find(class_='price large').get_text().strip("$").replace(',', ''))
    except AttributeError:
        da_price = ""

    try:
        # Need to get around no eight columns 'NoneType' Object
        da_desc = soup.find(class_='eight columns').get_text().strip()  # grab description
    except AttributeError:
        da_desc = soup.find(id='moredetailsTab').get_text().strip()

    da_pic = soup.find(class_='panel radius').findChild()['href']  # grab pic info
    da_upc = "UPC/Barcode" in soup.find(id='itemdetailsTab').find_all("li")[3].get_text()
    da_upc_text = ""
    da_stock_text = ""

    # logic for barcode information
    if da_upc:
        da_upc_text = soup.find(id='itemdetailsTab').find_all("li")[3].get_text().strip("UPC/Barcode:")

    # logic for stock information
    if da_price == "":
        da_stock_text = "Out Of Stock"
    else:
        da_stock_text = "In Stock"

    return [[title, da_price, da_desc, da_pic, da_upc_text, da_stock_text]]


# TODO: fix missing field issues
def parse_tcg(r):
    tcg_title_field = r.html.find(".product-details__name", first=True)  # Grabs item Name
    if tcg_title_field is None:
        print("     Error: title not found in html")
        tcg_title = ""
    else:
        tcg_title = tcg_title_field.text

    tcg_price_field = r.html.find(".spotlight__price", first=True)  # Grabs price
    if tcg_price_field is None:
        print("    Error: price not found in html")
        tcg_price = ""
    else:
        tcg_price = tcg_price_field.text

    tcg_desc_field = r.html.find(".pd-description__description", first=True)  # Grabs Description
    if tcg_desc_field is None:
        print("    Error: description not found in html")
        tcg_desc = ""
    else:
        tcg_desc = tcg_desc_field.text

    tcg_pic_field = r.html.find(".product-details__product", first=True)  # Grabs main picture
    if tcg_pic_field is None:
        print("    Error: pic field not found in html")
        tcg_pic = ""
    else:
        try:
            tcg_pic = tcg_pic_field.find(".progressive-image-main", first=True).attrs["src"]
        except KeyError:
            print("    Error: source link not found in pic field")
            tcg_pic = ""

    # logic for stock information
    tcg_stock_text = ""
    if tcg_price == "":
        tcg_stock_text = "Out Of Stock"
    else:
        tcg_stock_text = "In Stock"

    return [[tcg_title, tcg_price, tcg_desc, tcg_pic, tcg_stock_text]]


def parse_tnt(soup):
    tnt_title = soup.find('h1').get_text().strip()
    tnt_pics = soup.find(id='main-prod-img').findChild()['data-src']

    try:
        tnt_price = float(soup.find(class_="d-flex flex-column").span.get_text().strip('$'))
    except AttributeError:
        tnt_price = ""

    tnt_stock_text = ""

    # logic for stock information
    if tnt_price == "":
        tnt_stock_text = "Out of Stock"
    else:
        tnt_stock_text = "In Stock"

    return [[tnt_title, tnt_price, tnt_pics, tnt_stock_text]]


def parse_cs(soup):
    cs_title = soup.find('h1').get_text().strip()

    try:
        cs_price = float(soup.find(class_="price price--withoutTax").get_text().strip("$"))
    except AttributeError:
        cs_price = ""

    try:
        cs_pic = soup.find(class_='productView-img-container').findChild()['href']
    except AttributeError:
        cs_pic = ""

    # logic for stock information
    if cs_price == "":
        cs_stock_text = "Out of Stock"
    else:
        cs_stock_text = "In Stock"

    return [[cs_title, cs_price, cs_pic, cs_stock_text]]


def retry(func, param, max_retries):
    for i in range(0, max_retries+1):
        try:
            results = func(param)
        except Exception as ex:
            template = "An exception of type {0} occurred."
            message = template.format(type(ex).__name__)
        else:
            return results
    print("Function '", func.__name__, "' failed after", i+1, "attempts:", message)
    return None


def get_host(url):
    if "dacardworld.com" in url:
        return "DA"
    elif "tcgplayer.com" in url:
        return "TCG"
    elif "trollandtoad.com" in url:
        return "TNT"
    elif "collectorstore.com" in url:
        return "CS"
    else:
        print("Unknown host for url:", url)
        return "Unknown"


def get_product_pages(urls):
    page_dict = []
    for url in urls:
        if url:
            session = HTMLSession()
            host = get_host(url)
            print("Connecting to", url)
            response_html = retry(session.get, url, MAX_RETRIES)
            if response_html:
                try:
                    print("Rendering html")
                    response_html.html.render(timeout=100, retries=MAX_RETRIES)
                    page_dict.append({"host": host, "url": url, "html": response_html})
                except pyppeteer.errors.TimeoutError as ex:
                    print("Failed to render page for", url, "after", MAX_RETRIES, "attempts:", type(ex).__name__)
            session.close()
    return page_dict


if __name__ == "__main__":
    main()
