package com.silan.iot.nettestmachine;

import android.app.AlertDialog;
import android.content.Intent;
import android.os.Bundle;
import android.support.v4.app.Fragment;
import android.view.LayoutInflater;
import android.view.View;
import android.view.ViewGroup;
import android.widget.Button;

import android_serialport_api.sample.ConsoleActivity;
import android_serialport_api.sample.LoopbackActivity;
import android_serialport_api.sample.Sending01010101Activity;
import android_serialport_api.sample.SerialPortPreferences;

public class SerialPortMainMenuFragment extends Fragment {
    private static final String ARG_SECTION_NUMBER = "section_number";

    public static SerialPortMainMenuFragment newInstance(int sectionNumber) {
        SerialPortMainMenuFragment fragment = new SerialPortMainMenuFragment();
        Bundle args = new Bundle();
        args.putInt(ARG_SECTION_NUMBER, sectionNumber);
        fragment.setArguments(args);
        return fragment;
    }

    public SerialPortMainMenuFragment() {
    }

    @Override
    public View onCreateView(LayoutInflater inflater, ViewGroup container,
                             Bundle savedInstanceState) {
        View rootView = inflater.inflate(R.layout.main, container, false);
        final Button buttonSetup = (Button)rootView.findViewById(R.id.ButtonSetup);
        buttonSetup.setOnClickListener(new View.OnClickListener() {
            public void onClick(View v) {
                startActivity(new Intent(getActivity(), SerialPortPreferences.class));
            }
        });

        final Button buttonConsole = (Button)rootView.findViewById(R.id.ButtonConsole);
        buttonConsole.setOnClickListener(new View.OnClickListener() {
            public void onClick(View v) {
                startActivity(new Intent(getActivity(), ConsoleActivity.class));
            }
        });

        final Button buttonLoopback = (Button)rootView.findViewById(R.id.ButtonLoopback);
        buttonLoopback.setOnClickListener(new View.OnClickListener() {
            public void onClick(View v) {
                startActivity(new Intent(getActivity(), LoopbackActivity.class));
            }
        });

        final Button button01010101 = (Button)rootView.findViewById(R.id.Button01010101);
        button01010101.setOnClickListener(new View.OnClickListener() {
            public void onClick(View v) {
                startActivity(new Intent(getActivity(), Sending01010101Activity.class));
            }
        });

        final Button buttonAbout = (Button)rootView.findViewById(R.id.ButtonAbout);
        buttonAbout.setOnClickListener(new View.OnClickListener() {
            public void onClick(View v) {
                AlertDialog.Builder builder = new AlertDialog.Builder(getActivity());
                builder.setTitle("About");
                builder.setMessage("Serial Port API");
                builder.show();
            }
        });

        final Button buttonQuit = (Button)rootView.findViewById(R.id.ButtonQuit);
        buttonQuit.setOnClickListener(new View.OnClickListener() {
            public void onClick(View v) {
                getActivity().finish();
            }
        });

        return rootView;
    }
}
