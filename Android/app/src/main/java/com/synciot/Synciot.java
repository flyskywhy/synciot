package com.synciot;

import android.content.Context;
import android.content.res.AssetManager;
import android.util.Log;

import org.unix4j.Unix4j;
import org.unix4j.unix.grep.GrepOption;
import org.unix4j.unix.sed.SedOption;

import java.io.File;
import java.io.FileOutputStream;
import java.io.IOException;
import java.io.InputStream;
import java.text.SimpleDateFormat;
import java.util.Date;
import java.util.regex.Matcher;
import java.util.regex.Pattern;

/**
 * Created by lizheng on 15-11-4.
 */
public class Synciot {
    private static final String TAG = "Synciot";

    private static Context CallerCtx;

    private static String dataPath;
    private static final String STORAGE_PATH = "/sdcard";
    private static final String CLIENT_DIR = "synciot";
    private static final String CLIENT_PATH = STORAGE_PATH + "/" + CLIENT_DIR;
    private static final String SYNC_DIR = "sync";
    private static final String SYNC_PATH = CLIENT_PATH + "/" + SYNC_DIR;
    private static final String CONFIG_PATH = CLIENT_PATH + "/config";
    private static final String IN_PATH = SYNC_PATH + "/in";

    private static final String ASSETS_SYNCTHING = "syncthing";
    private static String syncthing;
    private static final String ASSETS_SERVER_DEVICE_ID_TXT = "server_device_id.txt";
    private static final String SERVER_DEVICE_ID_TXT = CONFIG_PATH + "/" + ASSETS_SERVER_DEVICE_ID_TXT;
    private static final String CONFIG_XML = CONFIG_PATH + "/config.xml";

    private static String server_device_id;
    private static final String SERVER_DEFAULT_FOLDER_DEVICE = new StringBuilder()
            .append("        <device id=\"SERVER_DEVICE_ID\"></device>")
            .toString();
    private static final String SERVER_DEVICE = new StringBuilder()
            .append("    <device id=\"SERVER_DEVICE_ID\" name=\"Server\" compression=\"metadata\" introducer=\"false\">\n")
            .append("        <address>dynamic</address>\n")
            .append("    </device>")
            .toString();

    private static final String prefix = "user".toUpperCase();
    private static final String suffix = ".synciot".toUpperCase();

    private static String device_id;
    private static String device_id_short;
    private static boolean runningBusiness = false;

    public static String getDevice_id() {
        return device_id;
    }

    public static String getDevice_id_short() {
        return device_id_short;
    }

    public static void start(Context ctx) {
        startSyncthing(ctx);
        startBusiness();
    }

    public static void stop() {
        stopBusiness();
        stopSyncthing();
    }

    private static void startSyncthing(Context ctx) {
        // To get root at the very beginning
        ShellInterface.isSuAvailable();

        CallerCtx = ctx;
        dataPath = "/data/data/" + CallerCtx.getApplicationContext().getPackageName();
        syncthing = dataPath + "/" + ASSETS_SYNCTHING;
        File file = new File(syncthing);
        if (!file.exists()) {
            extractAssets(CallerCtx, ASSETS_SYNCTHING, syncthing);
        }
        // To avoid extractAssets() before install Superuser, so we always chmod here,
        // not just after extractAssets() and never chmod.
        ShellInterface.runCommand("chmod 755 " + syncthing);

        mkdirSync();
        generateConfigXml();

        device_id = Unix4j.fromFile(CONFIG_XML)
                .grep("^        <device id=")
                .sed("s/^.*id=\"//")
                .sed("s/\">.*//")
                .toStringResult();

        device_id_short = Unix4j.fromString(device_id)
                .sed("s/-.*//")
                .toStringResult();

        file = new File(SERVER_DEVICE_ID_TXT);
        if (!file.exists()) {
            extractAssets(CallerCtx, ASSETS_SERVER_DEVICE_ID_TXT, SERVER_DEVICE_ID_TXT);
        }
        server_device_id = Unix4j.cat(SERVER_DEVICE_ID_TXT)
                .toStringResult();

        if (isOriginConfigXml()) {
            sedMisc2ConfigXml();
            sedSync2ConfigXml();
        }

        Pattern pattern = Pattern.compile(syncthing);
        String out = ShellInterface.getProcessOutput("ps syncthing");
        Matcher matcher = pattern.matcher(out);
        if (!matcher.find()) {
            new Thread(new Runnable() {
                public void run() {
                    ShellInterface.runCommand(syncthing
                            + " -no-browser -no-restart -gui-address=0.0.0.0:8384 -home="
                            + CONFIG_PATH);
                }
            }).start();
        }
    }

    private static void stopSyncthing() {
        Pattern pattern = Pattern.compile(syncthing);
        String out = ShellInterface.getProcessOutput("ps syncthing");
        Matcher matcher = pattern.matcher(out);
        if (matcher.find()) {
            String[] ps = out.replaceAll("USER.*root", "").split(" +");
            String pid = ps[1];
            ShellInterface.runCommand("kill " + pid);
        }
    }

    private static void startBusiness() {
        if (!runningBusiness) {
            new Thread(new Runnable() {
                public void run() {
                    runningBusiness = true;

                    for (; ; ) {
                        if (!runningBusiness) {
                            return;
                        }

                        try {
                            Thread.sleep(1000);
                        } catch (InterruptedException e) {
                            e.printStackTrace();
                        }

                        File dir = new File(IN_PATH);
                        if (dir.exists() && dir.isDirectory()) {
                            String[] child = dir.list();
                            if (child != null) {
                                for (int i = 0; i < child.length; i++) {
                                    String fileName = child[i];

                                    if (fileName.toUpperCase().startsWith(prefix)) {
                                        if (fileName.toUpperCase().endsWith(suffix)) {
                                            int p1 = fileName.indexOf(".");
                                            int p2 = fileName.indexOf(".", p1 + 1);
                                            int p3 = fileName.indexOf(".", p2 + 1);

                                            int inCount = Integer.parseInt(fileName.substring(p1 + 1, p2));
                                            if (inCount == (countFiles(dir) - 1)) { // user*.*.*.synciot itself was not counted by server, so `-1` here
                                                String userIdNum = fileName.substring(prefix.length(), p1);

                                                String cmd = fileName.substring(p2 + 1, p3);

                                                SimpleDateFormat time = new SimpleDateFormat("yyyyMMddHHmmss");
                                                String outPath = SYNC_PATH + "/" + time.format(new Date());
                                                ShellInterface.runCommand("mkdir -p " + outPath);

                                                Business.main(cmd, IN_PATH, outPath);

                                                ShellInterface.runCommand("mv " + IN_PATH + " " + outPath + "/");
                                                String outCount = String.valueOf(countFiles(new File(outPath)));
                                                ShellInterface.runCommand("touch " + outPath + "/out" + userIdNum + "." + outCount + ".synciot");

                                                Log.d(TAG, outPath);
                                            }
                                        }
                                    }
                                }
                            }
                        }
                    }
                }
            }).start();
        }
    }

    private static void stopBusiness() {
        runningBusiness = false;
    }

    private static int countFiles(java.io.File file) {
        int count = 0;
        java.io.File[] fileList = file.listFiles();
        for (int i = 0; i < fileList.length; i++) {
            if (fileList[i].isDirectory()) {
                count = count + countFiles(fileList[i]);
            } else {
                if (!fileList[i].getName().toLowerCase().startsWith(".syncthing.")) { // this synciot project hack ".syncthing." specially here
                    count++;
                }
            }
        }
        return count;
    }

    private static void sedSync2ConfigXml() {
        String server_default_folder_device = Unix4j.fromString(SERVER_DEFAULT_FOLDER_DEVICE)
                .sed(SedOption.substitute, "SERVER_DEVICE_ID", server_device_id)
                .toStringResult();
        Unix4j.fromFile(CONFIG_XML)
                .sed(SedOption.substitute, "id=\"default\" path=\".*\"", "id=\"" + device_id + "\" path=\"" + SYNC_PATH + "/\"")
                .sed(SedOption.append, "^        <device id=", server_default_folder_device)
                .toFile(CONFIG_XML + ".tmp");
        ShellInterface.runCommand("mv " + CONFIG_XML + ".tmp " + CONFIG_XML);
    }

    private static void sedMisc2ConfigXml() {
        String server_device = Unix4j.fromString(SERVER_DEVICE)
                .sed(SedOption.substitute, "SERVER_DEVICE_ID", server_device_id)
                .toStringResult();
        Unix4j.fromFile(CONFIG_XML)
                .sed(SedOption.append, "^    </device>", server_device)
                .sed("s/urAccepted>0/urAccepted>-1/")
                .sed("s/autoUpgradeIntervalH>12/autoUpgradeIntervalH>0/")
                .toFile(CONFIG_XML + ".tmp");
        ShellInterface.runCommand("mv " + CONFIG_XML + ".tmp " + CONFIG_XML);
    }

    private static boolean isOriginConfigXml() {
        final String count = Unix4j.fromFile(CONFIG_XML)
                .grep(GrepOption.count, "id=\"default\"")
                .toStringResult();
        return !("0".equals(count));
    }

    private static void generateConfigXml() {
        File file = new File(CONFIG_XML);
        if (!file.exists()) {
            ShellInterface.runCommand("mkdir -p " + CONFIG_PATH);
            ShellInterface.runCommand(syncthing + " -generate=" + CONFIG_PATH);
        }
    }

    private static void mkdirSync() {
        File file = new File(SYNC_PATH + "/.stfolder");
        if (!file.exists()) {
            ShellInterface.runCommand("mkdir -p " + SYNC_PATH);
            ShellInterface.runCommand("touch " + SYNC_PATH + "/.stfolder");
        }

        if (ShellInterface.isSuAvailable()) {
            Pattern pattern = Pattern.compile(CLIENT_DIR + "/" + SYNC_DIR + " tmpfs");
            String out = ShellInterface.getProcessOutput("mount");
            Matcher matcher = pattern.matcher(out);
            if (!matcher.find()) {
                ShellInterface.runCommand("mount -t tmpfs -o mode=0777 none " + SYNC_PATH);
                ShellInterface.runCommand("touch " + SYNC_PATH + "/.stfolder");
            }
        }
    }

    private static void extractAssets(Context ctx, String assets, String path) {
        File file = new File(path);
        if (!file.exists()) {
            try {
                AssetManager am = ctx.getApplicationContext().getAssets();
                InputStream is;
                is = am.open(assets);
                FileOutputStream out = new FileOutputStream(path);
                byte[] buffer = new byte[1024 * 64];
                int read = is.read(buffer);

                while (read >= 0) {
                    out.write(buffer, 0, read);
                    read = is.read(buffer);
                }

                out.close();
            } catch (IOException e) {
                e.printStackTrace();
            }
        }
    }
}
